import {
  Armchair,
  ArrowDownUp,
  Award,
  Sunrise,
  Sunset,
  type LucideIcon,
} from "lucide-react"

export type Factor = "start" | "end" | "gap" | "gpa" | "seats"

/** The ranking factors in display order, shared by the Weights panel and schedule stats. */
export const FACTORS: { key: Factor; label: string; icon: LucideIcon }[] = [
  { key: "start", label: "Start", icon: Sunrise },
  { key: "end", label: "End", icon: Sunset },
  { key: "gap", label: "Gaps", icon: ArrowDownUp },
  { key: "gpa", label: "GPA", icon: Award },
  { key: "seats", label: "Seats", icon: Armchair },
]
