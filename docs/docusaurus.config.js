// @ts-check
import {themes as prismThemes} from 'prism-react-renderer';

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: 'IONOS DynDNS Updater',
  tagline: 'Lightweight Dynamic DNS updater for IONOS domains',
  favicon: 'img/favicon.ico',

  future: {
    v4: true,
  },

  url: 'https://a-belhadj.github.io',
  baseUrl: '/ionos-ddns/',

  organizationName: 'a-belhadj',
  projectName: 'ionos-ddns',

  onBrokenLinks: 'throw',
  trailingSlash: false,

  markdown: {
    mermaid: true,
    hooks: {
      onBrokenMarkdownLinks: 'warn',
    },
  },
  themes: ['@docusaurus/theme-mermaid'],

  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      'classic',
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: {
          routeBasePath: '/',
          sidebarPath: './sidebars.js',
          editUrl: 'https://github.com/a-belhadj/ionos-ddns/edit/main/docs/',
        },
        blog: false,
        theme: {
          customCss: './src/css/custom.css',
        },
      }),
    ],
  ],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      colorMode: {
        respectPrefersColorScheme: true,
      },
      navbar: {
        title: 'IONOS DynDNS Updater',
        items: [
          {
            type: 'docSidebar',
            sidebarId: 'mainSidebar',
            position: 'left',
            label: 'Docs',
          },
          {
            href: 'https://github.com/a-belhadj/ionos-ddns',
            label: 'GitHub',
            position: 'right',
          },
        ],
      },
      footer: {
        style: 'dark',
        links: [
          {
            title: 'Docs',
            items: [
              {label: 'Getting Started', to: '/'},
              {label: 'Configuration', to: '/configuration'},
              {label: 'Deployment', to: '/deployment'},
            ],
          },
          {
            title: 'Links',
            items: [
              {
                label: 'GitHub',
                href: 'https://github.com/a-belhadj/ionos-ddns',
              },
              {
                label: 'Container Registry',
                href: 'https://github.com/a-belhadj/ionos-ddns/pkgs/container/ionos-ddns',
              },
              {
                label: 'IONOS API Docs',
                href: 'https://developer.hosting.ionos.com/docs/dns',
              },
            ],
          },
        ],
        copyright: `Copyright © ${new Date().getFullYear()} IONOS DynDNS Updater. Built with Docusaurus.`,
      },
      prism: {
        theme: prismThemes.github,
        darkTheme: prismThemes.dracula,
        additionalLanguages: ['bash', 'yaml', 'docker'],
      },
    }),
};

export default config;
