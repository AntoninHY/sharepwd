export type ExpirationOption = {
  label: string;
  value: string;
};

export const EXPIRATION_OPTIONS: ExpirationOption[] = [
  { label: "5 minutes", value: "5m" },
  { label: "30 minutes", value: "30m" },
  { label: "1 hour", value: "1h" },
  { label: "4 hours", value: "4h" },
  { label: "24 hours", value: "24h" },
  { label: "7 days", value: "168h" },
  { label: "30 days", value: "720h" },
];

export const VIEW_OPTIONS = [1, 2, 3, 5, 10, 25, 50, 100] as const;
