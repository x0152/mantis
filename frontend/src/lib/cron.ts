const DAY_NAMES = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']
const DAY_FULL = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday']
const MONTH_FULL = [
  'January', 'February', 'March', 'April', 'May', 'June',
  'July', 'August', 'September', 'October', 'November', 'December',
]

function pad2(n: number): string {
  return n.toString().padStart(2, '0')
}

function describeDayOfWeek(dow: string): string | null {
  if (dow === '*' || dow === '?') return null
  if (dow === '1-5' || dow === 'MON-FRI') return 'on weekdays'
  if (dow === '0,6' || dow === '6,0' || dow === 'SAT,SUN' || dow === 'SUN,SAT') return 'on weekends'
  if (/^[0-6]$/.test(dow)) return `on ${DAY_FULL[parseInt(dow, 10)]}`
  if (/^[0-6](,[0-6])+$/.test(dow)) {
    const days = dow.split(',').map(d => DAY_NAMES[parseInt(d, 10)])
    return `on ${days.join(', ')}`
  }
  return null
}

function describeDayOfMonth(dom: string): string | null {
  if (dom === '*' || dom === '?') return null
  if (/^\d+$/.test(dom)) {
    const n = parseInt(dom, 10)
    const suffix = n === 1 || n === 21 || n === 31 ? 'st'
                 : n === 2 || n === 22 ? 'nd'
                 : n === 3 || n === 23 ? 'rd'
                 : 'th'
    return `on the ${n}${suffix}`
  }
  return null
}

function describeMonth(mon: string): string | null {
  if (mon === '*' || mon === '?') return null
  if (/^\d+$/.test(mon)) {
    const n = parseInt(mon, 10)
    if (n >= 1 && n <= 12) return `in ${MONTH_FULL[n - 1]}`
  }
  return null
}

export function describeCron(expr: string): string {
  if (!expr) return ''
  const parts = expr.trim().split(/\s+/)
  if (parts.length !== 5) return expr

  const [min, hour, dom, mon, dow] = parts

  if (min === '*' && hour === '*' && dom === '*' && mon === '*' && dow === '*') {
    return 'Every minute'
  }

  const stepMinMatch = min.match(/^\*\/(\d+)$/)
  if (stepMinMatch && hour === '*' && dom === '*' && mon === '*' && dow === '*') {
    const n = parseInt(stepMinMatch[1], 10)
    return n === 1 ? 'Every minute' : `Every ${n} minutes`
  }

  const stepHourMatch = hour.match(/^\*\/(\d+)$/)
  if (/^\d+$/.test(min) && stepHourMatch && dom === '*' && mon === '*' && dow === '*') {
    const n = parseInt(stepHourMatch[1], 10)
    return n === 1 ? `Hourly at :${pad2(parseInt(min, 10))}` : `Every ${n} hours at :${pad2(parseInt(min, 10))}`
  }

  if (/^\d+$/.test(min) && hour === '*' && dom === '*' && mon === '*' && dow === '*') {
    return `Hourly at :${pad2(parseInt(min, 10))}`
  }

  if (/^\d+$/.test(min) && /^\d+$/.test(hour)) {
    const time = `${pad2(parseInt(hour, 10))}:${pad2(parseInt(min, 10))} UTC`
    const dowDesc = describeDayOfWeek(dow)
    const domDesc = describeDayOfMonth(dom)
    const monDesc = describeMonth(mon)

    if (dom === '*' && mon === '*' && (!dowDesc || dow === '*')) {
      return `Daily at ${time}`
    }
    if (dowDesc && dom === '*' && mon === '*') {
      return `${capitalize(dowDesc)} at ${time}`.replace(/^On weekdays/, 'Weekdays')
        .replace(/^On weekends/, 'Weekends')
    }
    if (domDesc && mon === '*' && dow === '*') {
      return `Monthly ${domDesc} at ${time}`
    }
    if (domDesc && monDesc && dow === '*') {
      return `Yearly ${monDesc} ${domDesc} at ${time}`
    }
  }

  return expr
}

function capitalize(s: string): string {
  return s.charAt(0).toUpperCase() + s.slice(1)
}
