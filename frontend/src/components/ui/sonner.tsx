import { Toaster as Sonner } from "sonner"

type ToasterProps = React.ComponentProps<typeof Sonner>

const Toaster = ({ ...props }: ToasterProps) => {
  return (
    <Sonner
      className="toaster group"
      toastOptions={{
        classNames: {
          toast:
            "group toast group-[.toaster]:bg-white dark:group-[.toaster]:bg-zinc-900 group-[.toaster]:text-zinc-900 dark:group-[.toaster]:text-zinc-100 group-[.toaster]:border-zinc-200 dark:group-[.toaster]:border-zinc-800 group-[.toaster]:shadow-lg",
          description: "group-[.toast]:text-zinc-500",
          actionButton:
            "group-[.toast]:bg-teal-600 group-[.toast]:text-white",
          cancelButton:
            "group-[.toast]:bg-zinc-200 dark:group-[.toast]:bg-zinc-800 group-[.toast]:text-zinc-600 dark:group-[.toast]:text-zinc-400",
          success: "group-[.toaster]:text-emerald-400",
          error: "group-[.toaster]:text-red-400",
        },
      }}
      {...props}
    />
  )
}

export { Toaster }
