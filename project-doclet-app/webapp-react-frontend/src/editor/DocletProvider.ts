import * as Y from 'yjs'
import {
  Awareness,
  applyAwarenessUpdate,
  encodeAwarenessUpdate,
} from 'y-protocols/awareness'
import { base64ToBytes, bytesToBase64 } from '../utils'

export type ProviderOptions = {
  documentId: string
  clientId: string
  wsUrl: string
  doc: Y.Doc
  user: { name: string; color: string }
  onStatus?: (status: 'connected' | 'disconnected') => void
  onUserName?: (clientId: string, name: string) => void
}

type SocketMessage = {
  type: string
  document_id: string
  client_id: string
  payload: string
}

export class DocletProvider {
  awareness: Awareness

  private doc: Y.Doc
  private ws: WebSocket | null = null
  private documentId: string
  private clientId: string
  private wsUrl: string
  private onStatus?: (status: 'connected' | 'disconnected') => void
  private onUserName?: (clientId: string, name: string) => void
  private snapshotTimer: number | null = null

  constructor(options: ProviderOptions) {
    this.doc = options.doc
    this.documentId = options.documentId
    this.clientId = options.clientId
    this.wsUrl = options.wsUrl
    this.onStatus = options.onStatus
    this.onUserName = options.onUserName
    this.awareness = new Awareness(this.doc)
    this.awareness.setLocalStateField('user', {
      ...options.user,
      clientId: this.clientId,
    })

    this.doc.on('update', this.handleDocUpdate)
    this.awareness.on('update', this.handleAwarenessUpdate)
    this.connect()
  }

  updateUser(user: { name: string; color: string }) {
    this.awareness.setLocalStateField('user', {
      ...user,
      clientId: this.clientId,
    })
  }

  private connect() {
    const url = new URL(this.wsUrl)
    url.searchParams.set('document_id', this.documentId)
    url.searchParams.set('client_id', this.clientId)

    this.ws = new WebSocket(url.toString())
    this.ws.onopen = () => {
      this.onStatus?.('connected')
      this.sendSnapshot()
    }
    this.ws.onclose = () => {
      this.onStatus?.('disconnected')
    }
    this.ws.onmessage = (event) => {
      this.handleMessage(event.data)
    }
  }

  private handleMessage(raw: string) {
    let msg: SocketMessage
    try {
      msg = JSON.parse(raw) as SocketMessage
    } catch {
      return
    }
    if (!msg || msg.document_id !== this.documentId) {
      return
    }
    if (msg.type === 'user_name') {
      if (msg.payload) {
        this.onUserName?.(msg.client_id, msg.payload)
      }
      return
    }
    if (msg.client_id === this.clientId) {
      return
    }
    if (msg.type === 'yjs_update') {
      const update = base64ToBytes(msg.payload)
      Y.applyUpdate(this.doc, update, 'remote')
      return
    }
    if (msg.type === 'presence') {
      const update = base64ToBytes(msg.payload)
      applyAwarenessUpdate(this.awareness, update, 'remote')
    }
  }

  private handleDocUpdate = (update: Uint8Array, origin: unknown) => {
    if (origin === 'remote') {
      return
    }
    this.sendMessage('yjs_update', bytesToBase64(update))
    this.scheduleSnapshot()
  }

  private handleAwarenessUpdate = (
    { added, updated, removed }: { added: number[]; updated: number[]; removed: number[] }
  ) => {
    const changed = added.concat(updated).concat(removed)
    const update = encodeAwarenessUpdate(this.awareness, changed)
    this.sendMessage('presence', bytesToBase64(update))
  }

  private scheduleSnapshot() {
    if (this.snapshotTimer) {
      window.clearTimeout(this.snapshotTimer)
    }
    this.snapshotTimer = window.setTimeout(() => {
      this.sendSnapshot()
    }, 1500)
  }

  private sendSnapshot() {
    const update = Y.encodeStateAsUpdate(this.doc)
    this.sendMessage('yjs_snapshot', bytesToBase64(update))
  }

  private sendMessage(type: string, payload: string) {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      return
    }
    const msg: SocketMessage = {
      type,
      document_id: this.documentId,
      client_id: this.clientId,
      payload,
    }
    this.ws.send(JSON.stringify(msg))
  }

  destroy() {
    this.doc.off('update', this.handleDocUpdate)
    this.awareness.off('update', this.handleAwarenessUpdate)
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
    if (this.snapshotTimer) {
      window.clearTimeout(this.snapshotTimer)
    }
  }
}
