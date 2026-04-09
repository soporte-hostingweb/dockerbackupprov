# Variables de Email Disponibles para Plantillas de Factura

## Variables de Autologin

El módulo AccesoCF ahora proporciona las siguientes variables que pueden ser utilizadas en las plantillas de email de WHMCS:

### {$auto_login_link}
- **Descripción**: URL directa para el enlace de autologin
- **Uso**: Para enlaces de texto plano o como href en enlaces HTML
- **Ejemplo**: `{$auto_login_link}`

### {$auto_login_link_html}
- **Descripción**: Enlace HTML completo con estilos predefinidos
- **Uso**: Para botones o enlaces estilizados en emails
- **Ejemplo**: `{$auto_login_link_html}`

## Cómo Usar en Plantillas de Email

### 1. Acceder a las Plantillas
1. Ve a **Setup > Email Templates**
2. Selecciona el tipo **Invoice**
3. Edita la plantilla deseada (ej: "Invoice Created")

### 2. Insertar las Variables
En el campo de mensaje, puedes usar cualquiera de estas opciones:

#### Opción 1: Enlace Simple
```
Para pagar su factura, haga clic en el siguiente enlace:
{$auto_login_link}
```

#### Opción 2: Botón Estilizado
```
Para pagar su factura, haga clic en el siguiente botón:
{$auto_login_link_html}
```

#### Opción 3: Enlace Personalizado
```
<a href="{$auto_login_link}" style="background-color: #28a745; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px;">Pagar Ahora</a>
```

## Características del Enlace

- **Seguridad**: El enlace contiene un token único que expira automáticamente
- **Duración**: El enlace expira según la configuración del módulo (por defecto 20 días)
- **Acceso**: Permite al cliente acceder directamente a la factura sin necesidad de login
- **Restricciones**: Si está habilitado, solo permite acceso a la factura y páginas de pago

## Plantillas Compatibles

Estas variables están disponibles en todas las plantillas de tipo:
- **Invoice** (Facturas)
- **Invoice Reminder** (Recordatorios de factura)

## Configuración Adicional

El módulo también incluye:
- Botón de envío por WhatsApp en el área de administración
- Configuración de días de expiración
- Opción para restringir acceso solo a páginas de factura y pago
