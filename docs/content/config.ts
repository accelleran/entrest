import { defineCollection, z } from "astro:content";
import { docsSchema } from "@astrojs/starlight/schema";

export const collections = {
  docs: defineCollection({
    schema: docsSchema({
      extend: z.object({
        // global banner thanks to: https://hideoo.dev/notes/starlight-sitewide-banner
        banner: z.object({ content: z.string() }).default({
          content: `⚠️ This is a <strong>fork</strong> of <a href="https://github.com/lrstanley/entrest" target="_blank">lrstanley/entrest</a> with additional features. For the official documentation, visit <a href="https://lrstanley.github.io/entrest/" target="_blank">lrstanley.github.io/entrest</a>.`,
        }),
      }),
    }),
  }),
};
