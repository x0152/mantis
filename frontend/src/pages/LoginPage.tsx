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
        <div className="mb-8 flex items-center justify-center gap-3">
          <MantisLogo size={48} className="text-teal-500 dark:text-teal-400" />
          <div className="leading-tight">
            <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-100 tracking-tight">Mantis</h1>
            <p className="text-sm text-zinc-500 mt-0.5">Control Plane</p>
          </div>
        </div>

        <Card className="bg-white dark:bg-zinc-900/60 border-zinc-200 dark:border-zinc-800/80 p-6">
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

            {error && <p className="text-red-400 text-sm">{error}</p>}

            <Button
              type="submit"
              disabled={loading || !token.trim()}
              className="w-full py-2.5 bg-teal-500 hover:bg-teal-400 text-zinc-950 font-semibold text-sm"
            >
              {loading ? 'Signing in...' : 'Sign in'}
            </Button>
          </form>
        </Card>
      </div>
    </div>
  )
}
