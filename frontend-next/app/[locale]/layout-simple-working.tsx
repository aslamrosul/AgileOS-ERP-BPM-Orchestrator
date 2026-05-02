// This is the WORKING simple layout - keep as reference
import { NextIntlClientProvider } from 'next-intl';
import { locales } from '@/i18n';

export default async function LocaleLayout({
  children,
  params
}: {
  children: React.ReactNode;
  params: { locale: string };
}) {
  const { locale } = params;
  
  // Load messages
  const messages = (await import(`../../messages/${locale}.json`)).default;
  
  return (
    <html lang={locale}>
      <body>
        <NextIntlClientProvider locale={locale} messages={messages}>
          {children}
        </NextIntlClientProvider>
      </body>
    </html>
  );
}
