import { cn } from "@/lib/utils";

interface StatCardProps {
  label: string;
  value: string;
  caption: string;
  tone?: "default" | "success" | "danger";
  /** The one figure the screen leads with. */
  hero?: boolean;
  loading?: boolean;
}

const TONE_CLASSES: Record<NonNullable<StatCardProps["tone"]>, string> = {
  default: "text-text-primary",
  success: "text-success",
  danger: "text-danger",
};

// One cell of the summary strip. The strip draws the rules between cells.
export function StatCard({ label, value, caption, tone = "default", hero, loading }: StatCardProps) {
  return (
    <div className="min-w-0 px-4 py-4 first:pl-0">
      <p className="text-sm text-text-secondary">{label}</p>
      {loading ? (
        <div className={cn("mt-1 motion-safe:animate-pulse rounded bg-surface-2", hero ? "h-12 w-44" : "h-8 w-24")} />
      ) : (
        <p className={cn("font-display mt-1 tabular-nums", hero ? "text-5xl" : "text-2xl", TONE_CLASSES[tone])}>
          {value}
        </p>
      )}
      <p className="mt-0.5 text-xs text-text-muted">{caption}</p>
    </div>
  );
}
