export type ExpirationOption = {
  value: string;
};

export const EXPIRATION_OPTIONS: ExpirationOption[] = [
  { value: "5m" },
  { value: "30m" },
  { value: "1h" },
  { value: "4h" },
  { value: "24h" },
  { value: "168h" },
  { value: "720h" },
];

export const VIEW_OPTIONS = [1, 2, 3, 5, 10, 25, 50, 100] as const;
