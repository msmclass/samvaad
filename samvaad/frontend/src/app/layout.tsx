import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Samvaad Meet — Sovereign Bharat SFU',
  description: 'Completely free virtual meeting platform with zero trackers, military-grade E2EE, and low-resource optimizations.',
  icons: {
    icon: '/assets/icons/favicon.ico',
    shortcut: '/assets/icons/favicon-32x32.png',
    apple: '/assets/icons/apple-touch-icon.png',
  },
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <head>
        <style dangerouslySetInnerHTML={{ __html: `
          html, body {
            width: 100%;
            height: 100%;
            margin: 0;
            padding: 0;
            background-color: #0b0c0e;
            color: #f3f4f6;
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
            overflow: hidden;
          }
          /* Scrollbar styling */
          ::-webkit-scrollbar {
            width: 6px;
            height: 6px;
          }
          ::-webkit-scrollbar-track {
            background: #121417;
          }
          ::-webkit-scrollbar-thumb {
            background: #272a30;
            border-radius: 3px;
          }
          ::-webkit-scrollbar-thumb:hover {
            background: #3e444f;
          }
        `}} />
      </head>
      <body>
        {children}
      </body>
    </html>
  );
}
