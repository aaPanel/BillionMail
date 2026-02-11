// Provide a standards-compliant localStorage for tests
// happy-dom v20 has a Proxy-based localStorage that causes issues with
// @vue/devtools-kit and direct localStorage.clear() calls
const store = new Map<string, string>()

const localStorageMock: Storage = {
  getItem: (key: string) => store.get(key) ?? null,
  setItem: (key: string, value: string) => { store.set(key, String(value)) },
  removeItem: (key: string) => { store.delete(key) },
  clear: () => { store.clear() },
  get length() { return store.size },
  key: (index: number) => [...store.keys()][index] ?? null,
}

Object.defineProperty(globalThis, 'localStorage', {
  value: localStorageMock,
  writable: true,
  configurable: true,
})
