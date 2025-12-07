# Nuxt 4 Full Conversion Guide

## Overview

TCE has been completely rewritten as a **Nuxt 4** application with **Nitro** server. All functionality has been ported from Go to Node.js/TypeScript:
- DOCX parsing with mammoth.js
- Document analysis engine in TypeScript
- YAML rule loading and application
- OpenAI integration for LLM suggestions

Single integrated application. No separate backend needed.

## Architecture

### Previous Stack (Go + Vue.js)
```
Frontend (Vue.js)  →  Go Backend  →  YAML Rules + OpenAI
Port 3000              Port 8080
```

### New Stack (Nuxt 4 Only)
```
Nuxt 4 Application (Single Port)
├── Frontend (Vue 3)
└── Nitro Server (Node.js)
    ├── DOCX Parser (mammoth.js)
    ├── Analyzer Engine (TypeScript)
    ├── Rule Loader (YAML)
    └── LLM Handler (OpenAI)
```

## What's New

### Complete Rewrite
- **DOCX Parser**: Now uses `mammoth.js` instead of Go's `unioffice`
- **Analyzer Engine**: Ported from Go to TypeScript
- **Rule System**: YAML rules still used, now loaded by Node.js
- **LLM Integration**: Direct OpenAI SDK in Nitro

### Key Features Preserved
- Multi-language support (Lithuanian/English)
- Study program selection (PI/SE)
- Real-time document analysis
- AI-powered text suggestions
- Interactive 3-column UI
- Comment categorization and filtering

## Project Structure

```
/tce/
├── app.vue                    # Root Nuxt component
├── pages/
│   └── index.vue              # Main application page
├── server/
│   ├── api/tce/
│   │   ├── health.ts          # Health check endpoint
│   │   ├── check.ts           # Document analysis
│   │   └── suggest-rewrite.ts # LLM suggestions
│   └── utils/
│       ├── types.ts           # TypeScript interfaces
│       ├── docparser.ts       # DOCX parsing (mammoth)
│       ├── config.ts          # YAML rule loading
│       ├── analyzer.ts        # Analysis engine
│       └── llm.ts             # OpenAI integration
├── config/                    # YAML rule files (unchanged)
├── public/                    # Static assets
├── nuxt.config.ts            # Nuxt configuration
├── tsconfig.json             # TypeScript config
├── package.json              # Node.js dependencies
└── .env.example              # Environment template
```

## Setup & Running

### Prerequisites

1. **Node.js** (v18+ recommended)
2. **OpenAI API Key** (for LLM suggestions)

### Installation

1. Clone and navigate to project:
   ```bash
   cd tce
   ```

2. Copy environment template:
   ```bash
   cp .env.example .env
   ```

3. Update `.env` with your OpenAI API key:
   ```env
   OPENAI_API_KEY=sk-your-key-here
   PORT=3000
   NODE_ENV=development
   ```

4. Install dependencies:
   ```bash
   npm install
   ```

### Development

Start the development server:
```bash
npm run dev
```

Open `http://localhost:3000` in your browser.

Features:
- Hot Module Replacement (HMR) for instant updates
- Server-side rendering (SSR) enabled
- File watching for changes

### Production Build

Build for production:
```bash
npm run build
```

Preview the build:
```bash
npm run preview
```

The production build creates optimized output in `.output/` directory ready for deployment.

## API Endpoints

All endpoints are local (no external backend):

### `GET /api/tce/health`
Health check
```json
{
  "status": "ok",
  "engine": "TCE",
  "version": "1.0.0",
  "runtime": "Nuxt 4 + Nitro"
}
```

### `POST /api/tce/check`
Document analysis

**Request:**
```
Form data:
- file: DOCX document
- lang: "lt" or "en"
- program: "pi" or "se"
```

**Response:**
```json
{
  "sections": [...],
  "paragraphs": [...],
  "contentBlocks": [...],
  "comments": [...]
}
```

### `POST /api/tce/suggest-rewrite`
Generate improved text with OpenAI

**Request:**
```json
{
  "paragraphText": "...",
  "commentTitle": "...",
  "commentMessage": "...",
  "lang": "lt"
}
```

**Response:**
```json
{
  "suggestion": "Improved text here..."
}
```

## Core Components

### Document Parser (`server/utils/docparser.ts`)
- Parses DOCX files using mammoth.js
- Extracts text, sections, and paragraphs
- Detects academic document structure
- Creates content blocks for UI rendering

### Analyzer Engine (`server/utils/analyzer.ts`)
- Applies YAML rules to parsed documents
- Checks structure, language, content, formatting
- Generates comments with severity levels
- Categories: structure, content, language, formatting

### Rule Loader (`server/utils/config.ts`)
- Loads YAML rule files from `/config/` directory
- Caches rules for performance
- Supports different programs (PI, SE) and languages (LT, EN)

### LLM Handler (`server/utils/llm.ts`)
- Integrates with OpenAI API
- Generates academic text suggestions
- Supports Lithuanian and English prompts
- Returns improved paragraphs addressing specific issues

## Environment Variables

| Variable | Default | Purpose |
|----------|---------|---------|
| `PORT` | `3000` | Server port |
| `NODE_ENV` | `development` | Node environment |
| `OPENAI_API_KEY` | (required) | OpenAI API key |
| `OPENAI_MODEL` | `gpt-4o-mini` | LLM model to use |

## Frontend Integration

The Vue.js component (`pages/index.vue`) features:

**Functionality:**
- Document upload with drag-drop support
- Language/program selection
- Real-time analysis results
- Interactive document viewer
- Comment filtering and navigation
- AI text suggestions per paragraph

**UI Layout:**
- Left sidebar: Summary & statistics
- Center: Document content viewer
- Right sidebar: Issues & comments

## Development Tips

### Adding Rules

YAML rule files are in `/config/`:
- `rules_pi_lt.yaml` - Practical Informatics (Lithuanian)
- `rules_pi_en.yaml` - Practical Informatics (English)
- `rules_se_lt.yaml` - Software Engineering (Lithuanian)
- `rules_se_en.yaml` - Software Engineering (English)

Rules are loaded at runtime and cached for performance.

### Extending the Analyzer

Edit `server/utils/analyzer.ts` to add new checks:

```typescript
private checkMyCustomRule(
  paragraphs: Paragraph[],
  comments: Comment[]
): void {
  // Your logic here
  comments.push({
    id: `comment-${nanoid(8)}`,
    paragraphId: 'my-para-id',
    category: 'content',
    severity: 'major',
    sectionLabel: 'My Section',
    title: 'Rule title',
    message: 'Rule description'
  })
}
```

Then call it from `analyze()` method.

### Debugging

Enable verbose logging:
```typescript
console.log('Debug info:', data)
```

Check server logs in terminal running `npm run dev`.

## Performance

- **Caching**: YAML rules cached in memory
- **Streaming**: Large file uploads handled with streams
- **Optimization**: Tree-shaking and code splitting in production build
- **SSR**: Enabled for faster first page load

## Troubleshooting

### "OPENAI_API_KEY is not set"
```bash
# Check your .env file
cat .env
# Make sure OPENAI_API_KEY is set with your actual key
```

### "Failed to parse DOCX file"
- Ensure the file is a valid DOCX document
- Check file size (should be < 50MB)
- Verify mammoth.js is installed: `npm list mammoth`

### "Port 3000 already in use"
```bash
# Change PORT in .env
PORT=3001
```

### LLM suggestions not working
- Verify OPENAI_API_KEY is correct
- Check OpenAI API account has credits
- Review server logs for errors

## Deployment

### Vercel
1. Push to GitHub
2. Connect repo to Vercel
3. Set environment variables in Vercel dashboard
4. Deploy (automatic on push)

### Netlify
```bash
npm run build
# Upload .output directory contents
```

### Traditional Server
```bash
npm run build
npm run preview
# Or use PM2: pm2 start ".output/server/index.mjs"
```

## Migration from Go Backend (if needed)

If you still need the old Go backend:
1. Keep Go binary and YAML config files
2. Old API calls go to `http://localhost:8080`
3. But this Nuxt app works standalone

## Testing

Run unit tests (when added):
```bash
npm run test
```

Generate coverage report:
```bash
npm run test:coverage
```

## What Happened to the Go Code?

The Go code is preserved in the repository but no longer used by Nuxt:
- `/cmd/` - Original Go server (can be removed if not needed)
- `/internal/` - Original Go packages
- `go.mod / go.sum` - Go dependencies

You can keep it for reference or to run separately if needed.

## Support & Resources

- [Nuxt 4 Documentation](https://nuxt.com/)
- [Nitro Server Guide](https://nitro.unjs.io/)
- [OpenAI API Documentation](https://platform.openai.com/docs)
- [Mammoth.js DOCX Parser](https://github.com/mwilliamson/mammoth.js)

## Summary

✅ Complete Nuxt 4 conversion
✅ All functionality ported to Node.js/TypeScript
✅ Single integrated application
✅ No separate backend needed
✅ Fully responsive UI
✅ Production-ready build system
✅ OpenAI integration for suggestions
