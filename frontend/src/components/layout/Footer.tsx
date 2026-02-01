'use client';

/**
 * Footer Component
 * 
 * Simple footer - good practice to have one on every site
 * Shows copyright, links, version info
 */

export default function Footer() {
  const currentYear = new Date().getFullYear();

  return (
    <footer className="bg-slate-900 border-t border-slate-700 py-6 px-4 mt-12">
      <div className="max-w-screen-2xl mx-auto flex flex-col md:flex-row items-center justify-between gap-4">
        
        {/* Left: Copyright */}
        <div className="text-slate-400 text-sm">
          <p>&copy; {currentYear} MogulMaker. All rights reserved.</p>
        </div>

        {/* Right: Links */}
        <div className="flex gap-6 text-slate-400 text-sm">
          <a href="#" className="hover:text-white transition">Privacy</a>
          <a href="#" className="hover:text-white transition">Terms</a>
          <a href="#" className="hover:text-white transition">Support</a>
        </div>
      </div>
    </footer>
  );
}
