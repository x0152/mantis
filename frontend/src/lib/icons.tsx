import { type ComponentType } from 'react'
import {
  Plus as PPlus,
  ChatText as PChatText,
  ChatCircle as PChatCircle,
  GitBranch as PGitBranch,
  GitFork as PGitFork,
  PencilSimple as PPencil,
  Trash as PTrash,
  Check as PCheck,
  CheckCircle as PCheckCircle,
  X as PX,
  XCircle as PXCircle,
  CaretDown as PCaretDown,
  CaretRight as PCaretRight,
  CircleNotch as PCircleNotch,
  Plug as PPlug,
  Scroll as PScroll,
  Sparkle as PSparkle,
  Broadcast as PBroadcast,
  ShieldWarning as PShieldWarning,
  Shield as PShield,
  ShieldCheck as PShieldCheck,
  Wrench as PWrench,
  SignOut as PSignOut,
  Cube as PCube,
  Cloud as PCloud,
  PaperPlaneTilt as PPaperPlane,
  Brain as PBrain,
  Square as PSquare,
  Paperclip as PPaperclip,
  FileText as PFileText,
  Image as PImage,
  Terminal as PTerminal,
  Calculator as PCalculator,
  DownloadSimple as PDownload,
  Microphone as PMic,
  Eye as PEye,
  MagicWand as PMagicWand,
  Play as PPlay,
  Pause as PPause,
  PauseCircle as PPauseCircle,
  Bell as PBell,
  Stack as PStack,
  CornersOut as PCornersOut,
  ArrowCounterClockwise as PArrowCCW,
  ArrowClockwise as PArrowCW,
  Copy as PCopy,
  WarningCircle as PWarningCircle,
  Sun as PSun,
  Moon as PMoon,
  Gauge as PGauge,
  ArrowLeft as PArrowLeft,
  FloppyDisk as PFloppy,
  Lightning as PLightning,
  Circle as PCircle,
  SkipForward as PSkipForward,
  Prohibit as PProhibit,
  Clock as PClock,
  Funnel as PFunnel,
  LinkSimple as PLinkSimple,
  Path as PPath,
  Wallet as PWallet,
  Infinity as PInfinity,
  type IconProps as PhosphorIconProps,
  type IconWeight,
} from '@phosphor-icons/react'

export type IconProps = Omit<PhosphorIconProps, 'ref'> & {
  strokeWidth?: number
}

export type IconComponent = ComponentType<IconProps>
export type LucideIcon = IconComponent

function adapt(C: ComponentType<PhosphorIconProps>, defaultWeight: IconWeight = 'regular'): IconComponent {
  const Wrapped: IconComponent = ({ strokeWidth, weight, size, ...rest }) => {
    void strokeWidth
    return <C size={size ?? 16} weight={weight ?? defaultWeight} {...rest} />
  }
  Wrapped.displayName = C.displayName ?? 'Icon'
  return Wrapped
}

export const Plus = adapt(PPlus)
export const MessageSquare = adapt(PChatCircle)
export const MessageSquareText = adapt(PChatText)
export const GitBranch = adapt(PGitBranch)
export const GitFork = adapt(PGitFork)
export const Pencil = adapt(PPencil)
export const Trash2 = adapt(PTrash)
export const Check = adapt(PCheck)
export const CheckCircle2 = adapt(PCheckCircle)
export const X = adapt(PX)
export const XCircle = adapt(PXCircle)
export const ChevronDown = adapt(PCaretDown)
export const ChevronRight = adapt(PCaretRight)
export const Loader2 = adapt(PCircleNotch, 'bold')
export const Plug = adapt(PPlug)
export const ScrollText = adapt(PScroll)
export const Sparkles = adapt(PSparkle)
export const Radio = adapt(PBroadcast)
export const ShieldAlert = adapt(PShieldWarning)
export const Shield = adapt(PShield)
export const ShieldCheck = adapt(PShieldCheck)
export const Wrench = adapt(PWrench)
export const LogOut = adapt(PSignOut)
export const Container = adapt(PCube)
export const Box = adapt(PCube)
export const Cloud = adapt(PCloud)
export const Send = adapt(PPaperPlane)
export const Brain = adapt(PBrain)
export const Square = adapt(PSquare, 'fill')
export const Paperclip = adapt(PPaperclip)
export const FileText = adapt(PFileText)
export const Image = adapt(PImage)
export const Terminal = adapt(PTerminal)
export const Calculator = adapt(PCalculator)
export const Download = adapt(PDownload)
export const Mic = adapt(PMic)
export const Eye = adapt(PEye)
export const Wand2 = adapt(PMagicWand)
export const Play = adapt(PPlay)
export const Pause = adapt(PPause)
export const PauseCircle = adapt(PPauseCircle)
export const Bell = adapt(PBell)
export const Layers = adapt(PStack)
export const Maximize2 = adapt(PCornersOut)
export const RotateCcw = adapt(PArrowCCW)
export const RotateCw = adapt(PArrowCW)
export const Copy = adapt(PCopy)
export const AlertCircle = adapt(PWarningCircle)
export const Sun = adapt(PSun)
export const Moon = adapt(PMoon)
export const Gauge = adapt(PGauge)
export const ArrowLeft = adapt(PArrowLeft)
export const Save = adapt(PFloppy)
export const Zap = adapt(PLightning)
export const Circle = adapt(PCircle)
export const SkipForward = adapt(PSkipForward)
export const Ban = adapt(PProhibit)
export const Clock = adapt(PClock)
export const Filter = adapt(PFunnel)
export const Link2 = adapt(PLinkSimple)
export const Route = adapt(PPath)
export const Wallet = adapt(PWallet)
export const Infinity = adapt(PInfinity)
