import { defineConfig } from 'astro/config';
import sitemap from '@astrojs/sitemap';

// No custom domain; the site metadata points at the GitHub repository.
export default defineConfig({
  site: 'https://github.com/naamfung/inx',
  build: { assets: 'static' },
  integrations: [sitemap({
    filter: (page) => !/\/changelog\/(?:stable|preview)\/?$/.test(page) &&
      !/\/changelog\/v\d+\.\d+\.\d+-/.test(page),
  })],
});
