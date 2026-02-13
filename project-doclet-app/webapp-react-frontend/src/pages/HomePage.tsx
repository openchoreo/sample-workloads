import { useEffect, useMemo, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { createDocument, listDocuments, DocumentListItem } from '../api'

export default function HomePage() {
  const navigate = useNavigate()
  const [query, setQuery] = useState('')
  const [docs, setDocs] = useState<DocumentListItem[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [displayName, setDisplayName] = useState('')

  const searchLabel = useMemo(() => (query ? `Results for "${query}"` : 'Recent documents'), [query])

  const refresh = async (nextQuery: string) => {
    setLoading(true)
    setError(null)
    try {
      const items = await listDocuments(nextQuery)
      setDocs(items)
    } catch (err) {
      setError((err as Error).message)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    refresh('')
  }, [])

  const onSearch = (event: React.FormEvent) => {
    event.preventDefault()
    refresh(query)
  }

  const onCreate = async (event: React.FormEvent) => {
    event.preventDefault()
    setLoading(true)
    setError(null)
    try {
      const doc = await createDocument(displayName)
      navigate(`/doc/${doc.document_id}`)
    } catch (err) {
      setError((err as Error).message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-white via-zinc-50 to-emerald-50/60">
      <div className="mx-auto max-w-5xl px-6 pb-8 pt-12">
        <div className="flex flex-wrap items-center justify-between gap-6">
          <div className="flex items-center gap-4">
            <img
              src="/doclet-logo-transparent.png"
              alt="Doclet mascot"
              className="h-16"
            />
            <div>
              <h1 className="text-4xl font-semibold tracking-tight text-zinc-900">Doclet</h1>
              <p className="mt-2 text-sm text-zinc-500">
                Anonymous real-time collaboration on rich-text documents
              </p>
            </div>
          </div>
        </div>

        <div className="mt-10 grid gap-10 lg:grid-cols-2">
          <div className="doclet-card p-6">
            <h2 className="text-lg font-semibold text-zinc-900">Create a new document</h2>
            <form onSubmit={onCreate} className="mt-5 flex flex-col gap-3">
              <input
                className="doclet-input"
                placeholder="Document name (optional)"
                value={displayName}
                onChange={(event) => setDisplayName(event.target.value)}
              />
              <button className="doclet-button" type="submit">Create document</button>
            </form>
          </div>
        </div>

        <div className="mt-10 doclet-card p-6">
          <div className="flex flex-col gap-4">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold text-zinc-900">{searchLabel}</h3>
              {loading ? <span className="text-sm text-zinc-500">Loading…</span> : null}
            </div>
            <form onSubmit={onSearch} className="flex flex-wrap gap-3">
              <input
                className="doclet-input flex-1"
                placeholder="Search by document name"
                value={query}
                onChange={(event) => setQuery(event.target.value)}
              />
              <button className="doclet-button-secondary" type="submit">Search</button>
            </form>
          </div>
          {error ? <div className="mt-3 text-sm text-rose-500">{error}</div> : null}
          {docs.length === 0 && !loading ? (
            <div className="mt-4 text-sm text-zinc-500">No documents found. Create one to get started!</div>
          ) : null}
          <div className="mt-4 grid gap-3">
            {docs.map((doc) => (
              <Link
                to={`/doc/${doc.document_id}`}
                key={doc.document_id}
                className="doclet-row flex items-center justify-between gap-4"
              >
                <div>
                  <div className="text-sm font-semibold text-zinc-900">
                    {doc.displayName || 'Untitled'}
                  </div>
                  <div
                    className="text-xs text-zinc-500"
                    title={new Date(doc.updated_at).toLocaleString()}
                  >
                    Updated {formatRelativeTime(doc.updated_at)}
                  </div>
                </div>
                <span className="text-sm font-semibold text-emerald-500">Open →</span>
              </Link>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}

function formatRelativeTime(value: string) {
  const updated = new Date(value).getTime()
  if (Number.isNaN(updated)) {
    return 'just now'
  }
  const diffMs = Date.now() - updated
  const diffSeconds = Math.floor(diffMs / 1000)
  if (diffSeconds < 60) {
    return 'just now'
  }
  const diffMinutes = Math.floor(diffSeconds / 60)
  if (diffMinutes < 60) {
    return `${diffMinutes}m ago`
  }
  const diffHours = Math.floor(diffMinutes / 60)
  if (diffHours < 24) {
    return `${diffHours}h ago`
  }
  const diffDays = Math.floor(diffHours / 24)
  if (diffDays < 7) {
    return `${diffDays}d ago`
  }
  return new Date(value).toLocaleDateString()
}
