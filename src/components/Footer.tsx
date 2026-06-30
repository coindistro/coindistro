'use client';

import { Shield } from 'lucide-react';

const footerLinks = {
  company: ['About Us', 'Careers', 'Blog', 'Press'],
  products: ['Exchange', 'Pay', 'Bank', 'Gift Cards', 'Academy', 'Signals', 'Bots', 'Invest'],
  resources: ['Documentation', 'API Reference', 'Status', 'Security'],
  legal: ['Privacy Policy', 'Terms of Service', 'Cookie Policy', 'Compliance'],
};

const socialLinks = ['X', 'Telegram', 'Discord', 'LinkedIn', 'YouTube'];

export default function Footer() {
  return (
    <footer className="relative py-16 border-t border-[var(--card-border)]">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-8 mb-12">
          {/* Brand */}
          <div className="lg:col-span-2">
            <div className="flex items-center gap-2 mb-4">
              <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-[#7C3AED] to-[#06B6D4] flex items-center justify-center">
                <Shield className="w-5 h-5 text-[var(--text-primary)]" />
              </div>
              <span className="text-xl font-bold gradient-text">Coindistro</span>
            </div>
            <p className="text-sm text-[var(--text-muted)] max-w-sm leading-relaxed mb-6">
              Africa's next-generation crypto financial ecosystem. Trade, learn, automate, and invest in one powerful platform.
            </p>
            {/* Social links */}
            <div className="flex items-center gap-4">
              {socialLinks.map((name) => (
                <a
                  key={name}
                  href="#"
                  className="text-sm text-[var(--text-muted)] hover:text-[var(--text-primary)] transition-colors duration-200"
                >
                  {name}
                </a>
              ))}
            </div>
          </div>

          {/* Links */}
          <div>
            <h4 className="text-sm font-semibold text-[var(--text-primary)] mb-4">Company</h4>
            <ul className="space-y-2">
              {footerLinks.company.map((link) => (
                <li key={link}>
                  <a href="#" className="text-sm text-[var(--text-muted)] hover:text-[var(--text-primary)] transition-colors duration-200">
                    {link}
                  </a>
                </li>
              ))}
            </ul>
          </div>

          <div>
            <h4 className="text-sm font-semibold text-[var(--text-primary)] mb-4">Products</h4>
            <ul className="space-y-2">
              {footerLinks.products.map((link) => (
                <li key={link}>
                  <a href="#" className="text-sm text-[var(--text-muted)] hover:text-[var(--text-primary)] transition-colors duration-200">
                    {link}
                  </a>
                </li>
              ))}
            </ul>
          </div>

          <div>
            <h4 className="text-sm font-semibold text-[var(--text-primary)] mb-4">Resources</h4>
            <ul className="space-y-2">
              {footerLinks.resources.map((link) => (
                <li key={link}>
                  <a href="#" className="text-sm text-[var(--text-muted)] hover:text-[var(--text-primary)] transition-colors duration-200">
                    {link}
                  </a>
                </li>
              ))}
            </ul>
          </div>
        </div>

        {/* Bottom */}
        <div className="pt-8 border-t border-[var(--card-border)] flex flex-col md:flex-row items-center justify-between gap-4">
          <p className="text-sm text-[var(--text-muted)]">
            © 2026 Coindistro. All rights reserved.
          </p>
          <div className="flex items-center gap-6">
            {footerLinks.legal.map((link) => (
              <a
                key={link}
                href="#"
                className="text-sm text-[var(--text-muted)] hover:text-[var(--text-primary)] transition-colors duration-200"
              >
                {link}
              </a>
            ))}
          </div>
        </div>
      </div>
    </footer>
  );
}
