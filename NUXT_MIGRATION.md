# Nuxt 4 + Nitro Migration Guide

## Overview

TCE has been migrated from a separate Go HTTP backend + Vue.js frontend to a unified **Nuxt 4** application with **Nitro** server engine. This provides a more integrated development experience while maintaining the existing Go backend for document analysis.

## Architecture

### Before (Separate Stack)
```
Frontend (Vue.js)  →  Go Backend
Port 3000              Port 8080
```

### After (Unified Stack)
```
Nuxt 4 Application (Frontend + Nitro API Proxy)
Port 3000
    ↓
  Nitro Server
    ↓
  Proxy Routes (/api/tce/*)
    ↓
  Go Backend (Document Analysis Engine)
  Port 8080
```

## Key Changes

### 1. Project Structure

**New Directories:**
- `/app.vue` - Root application component
- `/pages/` - Nuxt pages (currently index.vue for main app)
- `/components/` - Reusable Vue components
- `/server/` - Nitro server configuration
  - `/server/api/tce/` - API route handlers (proxy to Go backend)
- `/public/` - Static assets

**Updated Configuration:**
- `nuxt.config.ts` - Nuxt configuration with Nitro proxy settings
- `tsconfig.json` - TypeScript configuration
- `package.json` - Node.js dependencies

**Unchanged:**
- `/cmd/` - Go backend entry point (runs separately)
- `/internal/` - Go backend logic
- `/config/` - YAML rule files
- `go.mod / go.sum` - Go dependencies

### 2. Frontend Changes

The Vue.js frontend has been integrated into Nuxt:
- Moved from `/frontend/pages/index.vue` → `/pages/index.vue`
- Updated API calls to use relative paths (`/api/tce/*` instead of `http://localhost:8080/api/tce/*`)
- Uses `$fetch` (Nuxt's built-in composable) for API calls
- Removed dependency on external UI component libraries (using vanilla HTML + Tailwind CSS)

### 3. API Integration

Nitro API routes in `/server/api/tce/`:
- **`health.ts`** - Health check endpoint
- **`check.ts`** - Document analysis endpoint (proxies file uploads to Go backend)
- **`suggest-rewrite.ts`** - LLM rewrite suggestions endpoint

These routes:
- Forward requests to the Go backend
- Handle request/response transformation if needed
- Provide error handling and logging

## Running the Application

### Prerequisites

1. **Node.js** (v18+ recommended)
2. **Go backend** running on port 8080

### Setup

1. Copy `.env.example` to `.env`:
   ```bash
   cp .env.example .env
   ```

2. Update `GO_BACKEND_URL` in `.env` if your Go backend is on a different URL

3. Install Node.js dependencies:
   ```bash
   npm install
   ```

### Development

**Terminal 1: Start the Go backend**
```bash
go run cmd/server/main.go
```
Backend will run on `http://localhost:8080`

**Terminal 2: Start the Nuxt development server**
```bash
npm run dev
```
Frontend will be available at `http://localhost:3000`

The Nitro server will automatically proxy API requests from `http://localhost:3000/api/tce/*` to your Go backend.

### Production Build

```bash
npm run build
npm run preview
```

This builds the Nuxt application and previews it locally. For deployment, use the output from `npm run build`.

## Environment Variables

| Variable | Default | Purpose |
|----------|---------|---------|
| `PORT` | `3000` | Nuxt/Nitro server port |
| `GO_BACKEND_URL` | `http://localhost:8080` | Go backend URL for API proxy |
| `NODE_ENV` | `development` | Node.js environment |

## API Routes

All API routes are prefixed with `/api/tce/` and are proxied to the Go backend:

### `GET /api/tce/health`
- Health check endpoint
- Response: `{ status: "ok", engine: "TCE" }`

### `POST /api/tce/check`
- Document analysis endpoint
- Request: Multipart form with `file`, `lang`, `program`
- Response: Analysis results (sections, paragraphs, comments)

### `POST /api/tce/suggest-rewrite`
- LLM rewrite suggestions
- Request: JSON with paragraph text and comment details
- Response: `{ suggestion: "..." }`

## Frontend Features

The migrated frontend maintains all original features:
- Document upload (drag-and-drop ready)
- Language selection (Lithuanian/English)
- Study program selection (PI/SE)
- Real-time document analysis
- Issue visualization and filtering
- AI-powered text suggestions (via OpenAI)
- Responsive 3-column layout (summary | document | comments)

## Development Tips

### Modifying API Routes

API routes are in `/server/api/tce/`. To add or modify:

```typescript
// server/api/tce/myroute.ts
export default defineEventHandler(async (event) => {
  // Access request data
  const body = await readBody(event)

  // Forward to Go backend
  const response = await fetch(`${goBackendUrl}/api/tce/myroute`, {
    method: 'POST',
    body: JSON.stringify(body)
  })

  return response.json()
})
```

### Adding Frontend Components

Create reusable components in `/components/`:

```vue
<!-- components/MyComponent.vue -->
<template>
  <div>Component content</div>
</template>

<script setup lang="ts">
// Component logic
</script>
```

Then use them in pages:
```vue
<template>
  <MyComponent />
</template>
```

### Building for Deployment

The production build creates a `/dist` directory with all static assets. Deploy the entire output to your hosting platform (Vercel, Netlify, etc.).

## Troubleshooting

### "Cannot GET /api/tce/check"
- Ensure Go backend is running on `http://localhost:8080`
- Check `GO_BACKEND_URL` in `.env`
- Verify the Go backend's port in its logs

### "Port 3000 already in use"
- Change `PORT` in `.env` to an available port
- Or kill the process using port 3000

### API requests fail with CORS errors
- Ensure CORS is enabled on the Go backend
- Check that both servers are accessible to each other

## Migration Benefits

1. **Single Port** - Frontend and API on the same port (no CORS issues)
2. **Integrated Development** - Better DX with Nuxt tooling
3. **Shared Codebase** - Frontend and backend config in one place
4. **Hot Module Replacement** - Faster development with HMR
5. **Built-in Optimization** - Nuxt handles code splitting and bundling
6. **Server Middleware** - Can add authentication, logging, etc. at Nitro level

## Next Steps (Optional Enhancements)

- Add Tailwind CSS for better styling
- Create reusable Vue components for sections
- Add unit tests with Vitest
- Configure SSR caching strategy
- Add authentication if needed
- Deploy to Vercel/Netlify for automatic deployments

## Support

For issues with the Nuxt 4 migration, refer to:
- [Nuxt 4 Documentation](https://nuxt.com/docs)
- [Nitro Server Documentation](https://nitro.unjs.io/)
- [Original project README](./README.md)
