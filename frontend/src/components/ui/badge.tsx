import { cva, type VariantProps } from "class-variance-authority"
import * as React from "react"

import { cn } from "@/lib/utils"

const badgeVariants = cva(
  "inline-flex items-center gap-1 rounded-full px-1.5 py-0.5 text-[11px] font-medium",
  {
    variants: {
      variant: {
        default: "bg-blue-500/15 text-blue-400",
        secondary: "bg-violet-500/15 text-violet-400",
        success: "bg-emerald-500/15 text-emerald-400",
        destructive: "bg-red-500/10 text-red-400",
        warning: "bg-amber-500/15 text-amber-400",
        outline: "bg-amber-500/10 text-amber-400 border border-amber-500/20 font-mono rounded-md",
        muted: "bg-zinc-800 text-zinc-600",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  },
)

export interface BadgeProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof badgeVariants> {}

function Badge({ className, variant, ...props }: BadgeProps) {
  return (
    <div className={cn(badgeVariants({ variant }), className)} {...props} />
  )
}

export { Badge, badgeVariants }
