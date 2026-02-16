const DEBUG = true

export function logAuth(...args: unknown[]) {
  if (DEBUG) console.log('[AUTH]', ...args)
}

export function logApi(...args: unknown[]) {
  if (DEBUG) console.log('[API]', ...args)
}
