export const getErrorMessage = (error: unknown, fallback: string): string => {
  if (!error || typeof error !== 'object') return fallback

  const candidate = error as {
    backendMessage?: unknown
    message?: unknown
    response?: { data?: { message?: unknown } }
  }

  const message = candidate.backendMessage
    || candidate.response?.data?.message
    || candidate.message

  return typeof message === 'string' && message.trim() ? message : fallback
}
