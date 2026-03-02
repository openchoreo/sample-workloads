export type AppConfig = {
  docServiceUrl: string
  collabWsUrl: string
}

let cachedConfig: AppConfig | null = null

export async function loadConfig(): Promise<AppConfig> {
  if (cachedConfig) {
    return cachedConfig
  }
  try {
    const res = await fetch('/config.json', { cache: 'no-store' })
    if (res.ok) {
      cachedConfig = (await res.json()) as AppConfig
      return cachedConfig
    }
  } catch {
    // Fall back to defaults below.
  }
  cachedConfig = {
    docServiceUrl: '/api/document-svc',
    collabWsUrl: '/ws/collab-svc',
  }
  return cachedConfig
}
