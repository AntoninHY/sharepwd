import type { MetadataRoute } from "next";

const SITE_URL = "https://sharepwd.io";
const locales = ["en", "fr"];

export default function sitemap(): MetadataRoute.Sitemap {
  const pages = ["", "/create", "/docs"];

  return pages.flatMap((page) =>
    locales.map((locale) => ({
      url: `${SITE_URL}/${locale}${page}`,
      lastModified: new Date(),
      changeFrequency: "monthly" as const,
      priority: page === "" ? 1 : page === "/create" ? 0.9 : 0.7,
    }))
  );
}
