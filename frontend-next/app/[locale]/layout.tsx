import '../globals.css';

export default function LocaleLayout({
  children,
  params
}: {
  children: React.ReactNode;
  params: { locale: string };
}) {
  return (
    <html lang={params.locale}>
      <head>
        <title>AgileOS - Enterprise BPM Platform</title>
        <meta name="description" content="Enterprise Business Process Management Platform" />
      </head>
      <body>
        {children}
      </body>
    </html>
  );
}
