type Props = {
  label: string;
  onClick: () => void;
  children: React.ReactNode;
  // Chevrons point along the reading direction, so they mirror with it. Any
  // icon that means "back" or "forward" has to flip; a copy icon must not.
  mirror?: boolean;
};

export function IconButton({ label, onClick, children, mirror }: Props) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-label={label}
      title={label}
      className="flex size-7 shrink-0 items-center justify-center rounded-lg transition hover:bg-raised hover:text-ink"
    >
      <svg
        width="14"
        height="14"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.9"
        strokeLinecap="round"
        strokeLinejoin="round"
        className={mirror ? "rtl:-scale-x-100" : undefined}
      >
        {children}
      </svg>
    </button>
  );
}
