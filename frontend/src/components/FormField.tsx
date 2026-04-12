import type { ReactNode } from "react"
import { Label } from "@/components/ui/label"

interface FormFieldProps {
  label: string
  children: ReactNode
  hint?: string
  error?: string
}

export function FormField({ label, children, hint, error }: FormFieldProps) {
  return (
    <div>
      <Label>{label}</Label>
      {children}
      {hint && <p className="text-[11px] text-zinc-500 dark:text-zinc-600 mt-1">{hint}</p>}
      {error && <p className="text-xs text-red-400 mt-1">{error}</p>}
    </div>
  )
}
