import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Task board for one person",
  description: "One-person task board backed by Postgres."
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
