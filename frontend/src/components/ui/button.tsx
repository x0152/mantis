import * as React from "react"
import { Slot } from "@radix-ui/react-slot"
import { cva, type VariantProps } from "class-variance-authority"

import { cn } from "@/lib/utils"

const buttonVariants = cva(
  "inline-flex items-center justify-center gap-1.5 whitespace-nowrap rounded-[5px] text-[12px] font-medium tracking-tight transition-[background,color,transform,box-shadow] duration-100 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-teal-500/40 focus-visible:ring-offset-0 disabled:pointer-events-none disabled:opacity-40 active:translate-y-[0.5px] [&_svg]:pointer-events-none [&_svg]:shrink-0",
  {
    variants: {
      variant: {
        default:
          "bg-teal-600 text-white hover:bg-teal-500 shadow-[inset_0_1px_0_rgba(255,255,255,0.18),0_1px_0_rgba(0,0,0,0.25)] dark:shadow-[inset_0_1px_0_rgba(255,255,255,0.12),0_1px_0_rgba(0,0,0,0.5)]",
        secondary:
          "bg-zinc-50 dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 text-zinc-700 dark:text-zinc-300 hover:text-zinc-900 dark:hover:text-zinc-100 hover:border-zinc-300 dark:hover:border-zinc-700 hover:bg-white dark:hover:bg-zinc-800/60 shadow-[0_1px_0_rgba(0,0,0,0.02)] dark:shadow-none",
        outline:
          "border border-zinc-300 dark:border-zinc-700 text-zinc-700 dark:text-zinc-300 hover:border-teal-500/60 hover:text-teal-600 dark:hover:text-teal-400 bg-transparent",
        destructive:
          "text-zinc-500 hover:text-rose-500 hover:bg-rose-500/10 rounded-[4px]",
        ghost:
          "text-zinc-500 hover:text-teal-600 dark:hover:text-teal-400 hover:bg-teal-500/10 rounded-[4px]",
        link:
          "text-teal-600 dark:text-teal-400 underline-offset-4 hover:underline px-0",
      },
      size: {
        default: "h-8 px-3",
        sm: "h-7 px-2.5 text-[11.5px]",
        icon: "size-7 rounded-[4px]",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  },
)

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {
  asChild?: boolean
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, asChild = false, ...props }, ref) => {
    const Comp = asChild ? Slot : "button"
    return (
      <Comp
        className={cn(buttonVariants({ variant, size, className }))}
        ref={ref}
        {...props}
      />
    )
  },
)
Button.displayName = "Button"

export { Button, buttonVariants }
