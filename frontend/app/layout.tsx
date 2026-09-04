import type { Metadata } from "next";
import { Geist, Geist_Mono, Outfit } from "next/font/google";
import Script from "next/script";
import "./globals.css";

// Browser extensions (MetaMask, dst) inject scripts into every page and can throw
// (misal auto-connect gagal) — app ini tidak pakai web3 sama sekali, tapi Next.js dev
// overlay menangkap error apapun di window, termasuk dari extension, dan menampilkannya
// seolah bug aplikasi. Filter berdasarkan asal chrome-extension://, jalan paling awal
// (beforeInteractive) supaya nangkep sebelum listener overlay Next sendiri terpasang.
const SUPPRESS_EXTENSION_ERRORS = `
(function () {
  function isExtensionNoise(source) {
    return typeof source === "string" && /chrome-extension:\\/\\/|moz-extension:\\/\\//.test(source);
  }
  window.addEventListener("error", function (e) {
    if (isExtensionNoise(e.filename) || isExtensionNoise(e.error && e.error.stack)) {
      e.stopImmediatePropagation();
    }
  }, true);
  window.addEventListener("unhandledrejection", function (e) {
    var reason = e.reason;
    var stack = reason && reason.stack;
    if (isExtensionNoise(stack) || isExtensionNoise(String(reason))) {
      e.stopImmediatePropagation();
    }
  }, true);
})();
`;

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

const outfit = Outfit({
  variable: "--font-outfit",
  subsets: ["latin"],
  weight: ["400", "500", "600", "700"],
});

export const metadata: Metadata = {
  title: "Shared Finance",
  description: "Multi-tenant shared finance tracking app",
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html
      lang="en"
      data-theme="dark"
      data-scroll-behavior="smooth"
      className={`${geistSans.variable} ${geistMono.variable} ${outfit.variable} h-full antialiased`}
    >
      <head>
        <Script id="suppress-extension-errors" strategy="beforeInteractive">
          {SUPPRESS_EXTENSION_ERRORS}
        </Script>
      </head>
      <body className="min-h-full flex flex-col bg-base-100 text-base-content">{children}</body>
    </html>
  );
}
