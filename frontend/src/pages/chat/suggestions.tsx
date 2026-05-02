import { Sparkles, Eye, Container, Gauge, Bell, GitBranch, Wrench, Zap, Play, Shield, type IconComponent } from '@/lib/icons'

export interface Suggestion {
  icon: IconComponent
  title: string
  prompt: string
}

export const SUGGESTIONS: Suggestion[] = [
  {
    icon: Sparkles,
    title: 'What can you do?',
    prompt: 'In a few short bullet points, tell me what kinds of things you can help me with. Give me 3 simple examples I could try right now.',
  },
  {
    icon: Eye,
    title: 'Screenshot a website',
    prompt: 'Take a screenshot of the Hacker News homepage and send me the picture.',
  },
  {
    icon: Container,
    title: 'Trending repos this week',
    prompt: 'Find the 5 most popular GitHub projects from the last 7 days and show them in a small table — name, what it does, star count, and a link.',
  },
  {
    icon: Gauge,
    title: 'Plot Bitcoin price',
    prompt: 'Get the Bitcoin price for the last 30 days, draw a simple line chart, and send me the picture.',
  },
  {
    icon: Bell,
    title: 'Daily news brief',
    prompt: 'Set up a daily reminder that runs every weekday morning around 9:00. It should grab the top 5 Hacker News stories, write a one-sentence summary for each, and send the digest to me. Save it and tell me when it will run next.',
  },
  {
    icon: GitBranch,
    title: 'Plan a quick tour',
    prompt: 'Make a small plan with three steps: 1) list the most useful things you can do, 2) list what tools are ready right now, 3) suggest one easy task I could try next. Then run the plan and show me what came out.',
  },
  {
    icon: Wrench,
    title: 'Set up a Rust workspace',
    prompt: 'I want to play with Rust. Set up a fresh environment with Rust and Cargo, then write and run a tiny hello-world program in it so I see it works.',
  },
  {
    icon: Zap,
    title: 'Is this site healthy?',
    prompt: 'Check if github.com is up. Tell me how fast it responds, the status, and whether its security certificate looks fine. Keep it short.',
  },
  {
    icon: Play,
    title: 'Video → GIF',
    prompt: 'When I attach a short video, turn the first 5 seconds into a small GIF and send it back. Just say "ok" when you are ready.',
  },
  {
    icon: Shield,
    title: 'Quick security scan',
    prompt: 'Do a quick, friendly security scan of scanme.nmap.org (it is a public test target made for this). Show me which ports are open and what is listening, in a short table.',
  },
]
