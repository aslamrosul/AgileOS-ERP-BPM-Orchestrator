# AgileOS Internationalization (i18n) Implementation Guide

## 🌍 Overview

AgileOS sekarang mendukung **3 bahasa**:
- 🇬🇧 **English** (en)
- 🇮🇩 **Bahasa Indonesia** (id)  
- 🇨🇳 **中文 / Mandarin** (zh)

Sistem i18n menggunakan **next-intl** untuk manajemen terjemahan yang efisien dan type-safe.

## 📁 Struktur File

```
frontend-next/
├── messages/
│   ├── en.json          # English translations
│   ├── id.json          # Indonesian translations
│   └── zh.json          # Mandarin translations
├── components/
│   └── LanguageSwitcher.tsx  # Language selector component
├── i18n.ts              # i18n configuration
├── middleware.ts        # Locale detection middleware
└── next.config.mjs      # Next.js config with i18n plugin
```

## 🚀 Quick Start

### 1. Install Dependencies
```bash
cd frontend-next
npm install next-intl
```

### 2. Access Localized Routes
```
http://localhost:3000/en/workflow    # English
http://localhost:3000/id/workflow    # Indonesian
http://localhost:3000/zh/workflow    # Mandarin
```

### 3. Use Language Switcher
- Klik icon **Globe (🌐)** di navbar
- Pilih bahasa yang diinginkan
- Halaman akan otomatis reload dengan bahasa baru

## 💻 Usage in Components

### Basic Usage
```typescript
import { useTranslations } from 'next-intl';

export default function MyComponent() {
  const t = useTranslations('common');
  
  return (
    <button>{t('save')}</button>  // "Save" / "Simpan" / "保存"
  );
}
```

### Nested Keys
```typescript
const t = useTranslations('task.status');

<span>{t('approved')}</span>  // "Approved" / "Disetujui" / "已批准"
```

### With Parameters
```typescript
const t = useTranslations('notification');

t('newTask', { count: 5 })  // "5 new tasks assigned"
```

## 🎨 Font Support for Mandarin

### Add Noto Sans SC for Chinese Characters

Update `app/layout.tsx`:

```typescript
import { Noto_Sans_SC } from 'next/font/google';

const notoSansSC = Noto_Sans_SC({
  subsets: ['chinese-simplified'],
  weight: ['400', '500', '700'],
  variable: '--font-noto-sans-sc',
});

export default function RootLayout({ children }) {
  return (
    <html className={`${notoSansSC.variable}`}>
      <body className="font-sans">
        {children}
      </body>
    </html>
  );
}
```

Update `tailwind.config.ts`:

```typescript
module.exports = {
  theme: {
    extend: {
      fontFamily: {
        sans: ['var(--font-noto-sans-sc)', 'system-ui', 'sans-serif'],
      },
    },
  },
};
```

## 🗄️ Database Support for Multilingual Content

### Schema Modification

Update SurrealDB schema untuk support multilingual fields:

```surql
-- Workflow table with multilingual support
DEFINE TABLE IF NOT EXISTS workflow SCHEMAFULL;

DEFINE FIELD IF NOT EXISTS name ON workflow TYPE object;
DEFINE FIELD IF NOT EXISTS name.en ON workflow TYPE string;
DEFINE FIELD IF NOT EXISTS name.id ON workflow TYPE string;
DEFINE FIELD IF NOT EXISTS name.zh ON workflow TYPE string;

DEFINE FIELD IF NOT EXISTS description ON workflow TYPE object;
DEFINE FIELD IF NOT EXISTS description.en ON workflow TYPE string;
DEFINE FIELD IF NOT EXISTS description.id ON workflow TYPE string;
DEFINE FIELD IF NOT EXISTS description.zh ON workflow TYPE string;
```

### Example Data

```surql
CREATE workflow:leave_approval SET
    name = {
        en: "Leave Approval",
        id: "Persetujuan Cuti",
        zh: "请假批准"
    },
    description = {
        en: "Workflow for employee leave requests",
        id: "Alur kerja untuk permintaan cuti karyawan",
        zh: "员工请假申请工作流"
    },
    version = "1.0.0",
    is_active = true,
    created_at = time::now();
```

### Query with Language Preference

```surql
-- Get workflow with specific language
SELECT 
    id,
    name.zh AS name,
    description.zh AS description,
    version
FROM workflow
WHERE is_active = true;
```

## 🔧 Backend API Support

### Accept-Language Header

Update Go backend untuk support language header:

```go
// middleware/language.go
package middleware

import (
    "github.com/gin-gonic/gin"
)

func LanguageMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Get language from Accept-Language header
        lang := c.GetHeader("Accept-Language")
        
        // Default to English if not specified
        if lang == "" {
            lang = "en"
        }
        
        // Store in context for handlers to use
        c.Set("language", lang)
        c.Next()
    }
}
```

### Localized Error Messages

```go
// errors/messages.go
package errors

var ErrorMessages = map[string]map[string]string{
    "unauthorized": {
        "en": "Unauthorized. Please login again.",
        "id": "Tidak terotorisasi. Silakan masuk kembali.",
        "zh": "未授权。请重新登录。",
    },
    "not_found": {
        "en": "Resource not found.",
        "id": "Sumber daya tidak ditemukan.",
        "zh": "未找到资源。",
    },
    "server_error": {
        "en": "Server error. Please try again later.",
        "id": "Kesalahan server. Silakan coba lagi nanti.",
        "zh": "服务器错误。请稍后再试。",
    },
}

func GetErrorMessage(key, lang string) string {
    if messages, ok := ErrorMessages[key]; ok {
        if msg, ok := messages[lang]; ok {
            return msg
        }
        return messages["en"] // Fallback to English
    }
    return "An error occurred"
}
```

### Usage in Handlers

```go
func (h *WorkflowHandler) GetWorkflow(c *gin.Context) {
    lang := c.GetString("language")
    
    workflow, err := h.db.GetWorkflow(id)
    if err != nil {
        c.JSON(404, gin.H{
            "error": errors.GetErrorMessage("not_found", lang),
        })
        return
    }
    
    // Return localized workflow data
    c.JSON(200, gin.H{
        "id": workflow.ID,
        "name": workflow.Name[lang],
        "description": workflow.Description[lang],
    })
}
```

## 📝 Adding New Languages

### Step 1: Create Translation File

```bash
# Create new language file
cp messages/en.json messages/fr.json  # French example
```

### Step 2: Update i18n.ts

```typescript
export const locales = ['en', 'id', 'zh', 'fr'] as const;

export const localeNames: Record<Locale, string> = {
  en: 'English',
  id: 'Bahasa Indonesia',
  zh: '中文',
  fr: 'Français',  // Add new language
};
```

### Step 3: Update Middleware

```typescript
// middleware.ts
export default createMiddleware({
  locales: ['en', 'id', 'zh', 'fr'],  // Add new locale
  defaultLocale: 'en',
});

export const config = {
  matcher: ['/', '/(id|en|zh|fr)/:path*'],  // Add to matcher
};
```

### Step 4: Translate Content

Edit `messages/fr.json` and translate all keys.

## 🎯 Translation Keys Structure

### Common Keys
```json
{
  "common": {
    "loading": "...",
    "save": "...",
    "cancel": "...",
    "delete": "..."
  }
}
```

### Navigation
```json
{
  "nav": {
    "dashboard": "...",
    "workflows": "...",
    "tasks": "..."
  }
}
```

### Task Status (Critical for BPM)
```json
{
  "task": {
    "status": {
      "pending": "Pending / Menunggu / 待处理",
      "approved": "Approved / Disetujui / 已批准",
      "rejected": "Rejected / Ditolak / 已拒绝"
    }
  }
}
```

## 🌐 Production Deployment

### Environment Variables

```env
# .env.production
NEXT_PUBLIC_DEFAULT_LOCALE=en
NEXT_PUBLIC_SUPPORTED_LOCALES=en,id,zh
```

### Azure Deployment

1. **Build with i18n support**:
```bash
npm run build
```

2. **Verify locales in build**:
```bash
ls .next/static/chunks/
# Should see locale-specific chunks
```

3. **Deploy to Azure**:
```bash
docker build -t agileos-frontend .
docker push your-registry.azurecr.io/agileos-frontend:latest
```

4. **Access localized routes**:
```
https://your-app.azurewebsites.net/en/
https://your-app.azurewebsites.net/id/
https://your-app.azurewebsites.net/zh/
```

## 📊 Translation Coverage

### Current Coverage

| Category | English | Indonesian | Mandarin |
|----------|---------|------------|----------|
| Common UI | ✅ 100% | ✅ 100% | ✅ 100% |
| Navigation | ✅ 100% | ✅ 100% | ✅ 100% |
| Auth | ✅ 100% | ✅ 100% | ✅ 100% |
| Workflow | ✅ 100% | ✅ 100% | ✅ 100% |
| Tasks | ✅ 100% | ✅ 100% | ✅ 100% |
| Analytics | ✅ 100% | ✅ 100% | ✅ 100% |
| Notifications | ✅ 100% | ✅ 100% | ✅ 100% |
| Errors | ✅ 100% | ✅ 100% | ✅ 100% |

**Total Keys**: 120+ translations per language

## 🔍 Testing i18n

### Manual Testing Checklist

- [ ] Language switcher appears in navbar
- [ ] All 3 languages selectable
- [ ] Page reloads with correct language
- [ ] Language preference persists after refresh
- [ ] Mandarin characters display correctly (no boxes/tofu)
- [ ] Task statuses translated correctly
- [ ] Error messages in correct language
- [ ] Date/time formats localized

### Automated Testing

```typescript
// __tests__/i18n.test.ts
import { render, screen } from '@testing-library/react';
import { NextIntlClientProvider } from 'next-intl';
import messages from '@/messages/zh.json';

test('renders Mandarin translations', () => {
  render(
    <NextIntlClientProvider locale="zh" messages={messages}>
      <MyComponent />
    </NextIntlClientProvider>
  );
  
  expect(screen.getByText('保存')).toBeInTheDocument();
});
```

## 🎓 Best Practices

### 1. Always Use Translation Keys
❌ Bad:
```typescript
<button>Save</button>
```

✅ Good:
```typescript
<button>{t('common.save')}</button>
```

### 2. Avoid Hardcoded Strings
❌ Bad:
```typescript
toast.success("Workflow saved successfully");
```

✅ Good:
```typescript
toast.success(t('success.saved'));
```

### 3. Use Descriptive Keys
❌ Bad:
```json
{
  "btn1": "Save",
  "msg1": "Success"
}
```

✅ Good:
```json
{
  "common.save": "Save",
  "success.workflowSaved": "Workflow saved successfully"
}
```

### 4. Keep Translations Consistent
Ensure same terms are translated consistently across all files.

### 5. Test with Real Content
Test UI with longest translations (usually German/Indonesian) to ensure layout doesn't break.

## 📚 Resources

- [next-intl Documentation](https://next-intl-docs.vercel.app/)
- [Unicode CLDR](http://cldr.unicode.org/) - Locale data
- [Google Fonts - Noto Sans SC](https://fonts.google.com/noto/specimen/Noto+Sans+SC)
- [i18n Best Practices](https://www.w3.org/International/questions/qa-i18n)

## ✅ Implementation Checklist

- [x] Install next-intl
- [x] Create translation files (en, id, zh)
- [x] Configure i18n.ts
- [x] Create LanguageSwitcher component
- [x] Update next.config.mjs
- [x] Create middleware for locale detection
- [ ] Add LanguageSwitcher to navbar
- [ ] Update components to use translations
- [ ] Add Noto Sans SC font for Mandarin
- [ ] Update database schema for multilingual content
- [ ] Add Accept-Language support in backend
- [ ] Test all languages
- [ ] Deploy to production

---

**Status**: ✅ i18n Infrastructure Complete

**Next Steps**:
1. Add LanguageSwitcher to WorkflowCanvas toolbar
2. Update login/register pages with translations
3. Test Mandarin display on ThinkPad T14
4. Add backend Accept-Language middleware

**Demo**: 
```
http://localhost:3000/zh/workflow  # 中文界面
http://localhost:3000/id/workflow  # Bahasa Indonesia
http://localhost:3000/en/workflow  # English
```
