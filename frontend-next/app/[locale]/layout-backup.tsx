// Backup of original layout
import { NextIntlClientProvider } from 'next-intl';
import { notFound } from 'next/navigation';
import { Inter, Noto_Sans_SC } from 'next/font/google';
import { Toaster } from 'sonner';
import { WebSocketProvider } from '@/components/WebSocketProvider';
import { locales } from '@/i18n';
import '../globals.css';

const inter = Inter({ subsets: ['latin'] });
const notoSansSC = Noto_Sans_SC({ 
  subsets: ['latin'],
  variable: '--font-noto-sans-sc',
});

export function generateStaticParams() {
  return locales.map((locale) => ({ locale }));
}

export default async function LocaleLayout({
  children,
  params
}: {
  children: React.ReactNode;
  params: { locale: string };
}) {
  const { locale } = params;
  
  // Validate locale
  if (!locales.includes(locale as any)) {
    notFound();
  }

  // Import messages directly
  let messages;
  try {
    messages = (await import(`../../messages/${locale}.json`)).default;
  } catch (error) {
    console.error(`Failed to load messages for locale: ${locale}`, error);
    notFound();
  }

  return (
    <html lang={locale} className={locale === 'zh' ? notoSansSC.variable : ''}>
      <body className={inter.className}>
        <NextIntlClientProvider messages={messages}>
          <WebSocketProvider>
            {children}
            <Toaster position="top-right" richColors />
          </WebSocketProvider>
        </NextIntlClientProvider>
      </body>
    </html>
  );
}
