import type { MetadataRoute } from "next";

export default function robots(): MetadataRoute.Robots {
  return {
    rules: [
      {
        userAgent: "*",
        allow: ["/", "/create", "/docs"],
        disallow: ["/s/", "/f/", "/v1/"],
      },
    ],
    sitemap: "https://sharepwd.io/sitemap.xml",
  };
}
