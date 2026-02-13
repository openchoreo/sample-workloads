export function base64ToBytes(input: string): Uint8Array {
  if (!input) {
    return new Uint8Array()
  }
  const binary = atob(input)
  const bytes = new Uint8Array(binary.length)
  for (let i = 0; i < binary.length; i += 1) {
    bytes[i] = binary.charCodeAt(i)
  }
  return bytes
}

export function bytesToBase64(bytes: Uint8Array): string {
  let binary = ''
  bytes.forEach((byte) => {
    binary += String.fromCharCode(byte)
  })
  return btoa(binary)
}

export function getSessionClientId(): string {
  const key = 'doclet_client_id'
  const existing = sessionStorage.getItem(key)
  if (existing) {
    return existing
  }
  const id = crypto.randomUUID()
  sessionStorage.setItem(key, id)
  return id
}

const nameKey = 'doclet_display_name'
const adjectives = [
  'Brisk',
  'Calm',
  'Clever',
  'Golden',
  'Mellow',
  'Quick',
  'Quiet',
  'Sharp',
  'Sunny',
  'Witty',
]
const nouns = [
  'Comet',
  'Falcon',
  'Harbor',
  'Lighthouse',
  'Meadow',
  'Orchard',
  'River',
  'Sparrow',
  'Summit',
  'Willow',
]

export function getSessionDisplayName(): string {
  const existing = sessionStorage.getItem(nameKey)
  if (existing) {
    return existing
  }
  return resetSessionDisplayName()
}

export function setSessionDisplayName(name: string) {
  if (!name) {
    sessionStorage.removeItem(nameKey)
    return
  }
  sessionStorage.setItem(nameKey, name)
}

export function resetSessionDisplayName(): string {
  const name = `${sample(adjectives)} ${sample(nouns)}`
  sessionStorage.setItem(nameKey, name)
  return name
}

function sample(values: string[]) {
  return values[Math.floor(Math.random() * values.length)]
}

export function colorFromSeed(seed: string): string {
  let hash = 0
  for (let i = 0; i < seed.length; i += 1) {
    hash = seed.charCodeAt(i) + ((hash << 5) - hash)
  }
  const hue = Math.abs(hash) % 360
  return `hsl(${hue}, 70%, 45%)`
}
