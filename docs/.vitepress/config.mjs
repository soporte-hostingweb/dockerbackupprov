import { defineConfig } from 'vitepress'

export default defineConfig({
  title: "HW Cloud Recovery",
  description: "Enterprise Disaster Recovery as a Service",
  themeConfig: {
    nav: [
      { text: 'Inicio', link: '/' },
      { text: 'Administrador', link: '/admin/' },
      { text: 'Usuario', link: '/user/' },
      { text: 'Guías', link: '/guides/' },
      { text: 'Referencia Técnica', link: '/reference/' }
    ],
    sidebar: {
      '/admin/': [
        {
          text: 'Nivel Administrador',
          items: [
            { text: 'Instalación y Setup', link: '/admin/' },
            { text: 'Configuración WHMCS', link: '/admin/whmcs' },
            { text: 'Análisis Financiero', link: '/admin/finanzas' }
          ]
        }
      ],
      '/user/': [
        {
          text: 'Nivel Usuario',
          items: [
            { text: 'Cómo Instalar', link: '/user/' },
            { text: 'Escenarios y Ejemplos', link: '/user/escenarios' },
            { text: 'Planes', link: '/user/planes' }
          ]
        }
      ],
      '/guides/': [
        {
          text: 'Guías Básicas',
          items: [
            { text: 'Backup Total y Parcial', link: '/guides/' },
            { text: 'Restore Wizard', link: '/guides/restore' }
          ]
        }
      ],
      '/reference/': [
        {
          text: 'Documentación Técnica',
          items: [
            { text: 'Arquitectura y Seguridad', link: '/reference/' },
            { text: 'Funciones y Endpoints', link: '/reference/endpoints' },
            { text: 'Análisis y Funciones Obsoletas', link: '/reference/analisis' }
          ]
        }
      ]
    },
    socialLinks: [
      { icon: 'github', link: 'https://github.com/soporte-hostingweb/dockerbackupprov' }
    ]
  }
})
