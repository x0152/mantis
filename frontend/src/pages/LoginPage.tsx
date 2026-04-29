import { useState, type FormEvent } from 'react'
import { api, UnauthorizedError } from '../api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card } from '@/components/ui/card'
import { FormField } from '@/components/FormField'
import { MantisLogo } from '@/components/MantisLogo'
import type { User } from '../types'

export default function LoginPage({ onLogin }: { onLogin: (user: User) => void }) {
  const [token, setToken] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const submit = async (e: FormEvent) => {
    e.preventDefault()
    if (!token.trim()) return
    setLoading(true)
    setError('')
    try {
      const user = await api.auth.login(token.trim())
      onLogin(user)
    } catch (err) {
      if (err instanceof UnauthorizedError) setError('Invalid token')
      else setError(err instanceof Error ? err.message : 'Login failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-zinc-50 dark:bg-zinc-950 flex items-center justify-center p-6">
      <div className="w-full max-w-sm">
        <div className="mb-6 flex items-center justify-center gap-3">
          <MantisLogo size={40} className="text-teal-500 dark:text-teal-400" />
          <div className="leading-tight">
            <h1 className="font-mono text-[20px] font-medium lowercase tracking-tight text-zinc-900 dark:text-zinc-50">mantis</h1>
            <p className="font-mono text-[11px] lowercase tracking-tight text-zinc-500 mt-0.5">v0 · control</p>
          </div>
        </div>

        <Card className="bg-white dark:bg-zinc-900/60 p-6">
          <div className="kicker mb-4 pb-3 border-b border-zinc-200/60 dark:border-zinc-800/60">
            <span className="kicker-num">·</span>
            <span>auth</span>
          </div>
          <form onSubmit={submit} className="space-y-4">
            <FormField label="Access token">
              <Input
                type="password"
                value={token}
                onChange={e => setToken(e.target.value)}
                placeholder="your token"
                autoFocus
              />
            </FormField>

            {error && <p className="font-mono text-[11px] lowercase text-rose-500">err: {error}</p>}

            <Button
              type="submit"
              disabled={loading || !token.trim()}
              className="w-full h-9"
            >
              {loading ? 'signing in...' : 'sign in'}
            </Button>
          </form>
        </Card>
      </div>
    </div>
  )
}
