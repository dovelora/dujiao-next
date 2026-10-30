export function isCatalogPath(path: string): boolean {
  return path === '/' || path === '/products' || path === '/blog' || path.startsWith('/categories/')
}

export function catalogCanonicalPath(path: string, query: Record<string, unknown>): string {
  if (!isCatalogPath(path)) return path
  const page = typeof query.page === 'string' ? query.page : ''
  if (!/^[1-9]\d*$/.test(page) || !Number.isSafeInteger(Number(page)) || Number(page) <= 1) return path
  return `${path}?page=${Number(page)}`
}
