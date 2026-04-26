export type Route =
  | { page: 'chat'; sessionId?: string }
  | { page: 'plans'; planId?: string }
  | { page: 'channels' }
  | { page: 'connections' }
  | { page: 'skills' }
  | { page: 'logs' }
  | { page: 'llm' }
  | { page: 'guard-profiles' }
  | { page: 'runtimes' }

export type PageId = Route['page']

const patterns: [RegExp, (m: RegExpMatchArray) => Route][] = [
  [/^\/chat\/(.+)$/, m => ({ page: 'chat', sessionId: m[1] })],
  [/^\/chat\/?$/, () => ({ page: 'chat' })],
  [/^\/plans\/(.+)$/, m => ({ page: 'plans', planId: m[1] })],
  [/^\/plans\/?$/, () => ({ page: 'plans' })],
  [/^\/channels\/?$/, () => ({ page: 'channels' })],
  [/^\/connections\/?$/, () => ({ page: 'connections' })],
  [/^\/skills\/?$/, () => ({ page: 'skills' })],
  [/^\/logs\/?$/, () => ({ page: 'logs' })],
  [/^\/llm\/?$/, () => ({ page: 'llm' })],
  [/^\/guard-profiles\/?$/, () => ({ page: 'guard-profiles' })],
  [/^\/runtimes\/?$/, () => ({ page: 'runtimes' })],
]

export function parseRoute(path = window.location.pathname): Route {
  for (const [re, fn] of patterns) {
    const m = path.match(re)
    if (m) return fn(m)
  }
  return { page: 'chat' }
}

export function routePath(route: Route): string {
  switch (route.page) {
    case 'chat': return route.sessionId ? `/chat/${route.sessionId}` : '/chat'
    case 'plans': return route.planId ? `/plans/${route.planId}` : '/plans'
    default: return `/${route.page}`
  }
}

export function navigate(route: Route) {
  const path = routePath(route)
  if (window.location.pathname !== path) {
    window.history.pushState(null, '', path)
  }
  window.dispatchEvent(new PopStateEvent('popstate'))
}
