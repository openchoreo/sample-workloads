import { loadConfig } from './config'

export type DocumentListItem = {
  document_id: string
  displayName: string
  updated_at: string
}

export type DocumentResponse = {
  document_id: string
  displayName: string
  content: string
  created_at: string
  updated_at: string
}

export async function listDocuments(query: string): Promise<DocumentListItem[]> {
  const config = await loadConfig()
  const url = new URL(`${config.docServiceUrl}/documents`)
  if (query) {
    url.searchParams.set('query', query)
  }
  const res = await fetch(url.toString())
  if (!res.ok) {
    throw new Error('Failed to load documents')
  }
  const data = await res.json()
  return data.items || []
}

export async function createDocument(displayName: string): Promise<DocumentResponse> {
  const config = await loadConfig()
  const res = await fetch(`${config.docServiceUrl}/documents`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ displayName }),
  })
  if (!res.ok) {
    throw new Error('Failed to create document')
  }
  return res.json()
}

export async function getDocument(documentId: string): Promise<DocumentResponse> {
  const config = await loadConfig()
  const res = await fetch(`${config.docServiceUrl}/documents/${documentId}`)
  if (!res.ok) {
    throw new Error('Document not found')
  }
  return res.json()
}

export async function updateDocumentTitle(documentId: string, displayName: string): Promise<void> {
  const config = await loadConfig()
  const res = await fetch(`${config.docServiceUrl}/documents/${documentId}/title`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ displayName }),
  })
  if (!res.ok) {
    throw new Error('Failed to update title')
  }
}

export async function deleteDocument(documentId: string): Promise<void> {
  const config = await loadConfig()
  const res = await fetch(`${config.docServiceUrl}/documents/${documentId}`, {
    method: 'DELETE',
  })
  if (!res.ok) {
    throw new Error('Failed to delete document')
  }
}

export async function getCollabWsUrl(): Promise<string> {
  const config = await loadConfig()
  return config.collabWsUrl || 'ws://localhost:8090/ws'
}
