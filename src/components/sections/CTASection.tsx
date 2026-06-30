'use client';

import { motion } from 'framer-motion';
import { ArrowRight, Calendar, Sparkles } from 'lucide-react';

export default function CTASection() {
  return (
    <section id="cta" className="relative py-24 overflow-hidden">
      {/* Background glow */}
      <div className="absolute inset-0 bg-gradient-to-r from-[#7C3AED]/10 to-[#06B6D4]/10" />
      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[800px] h-[800px] bg-[#7C3AED] rounded-full blur-[200px] opacity-20" />

      <div className="relative max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
        <motion.div
          initial={{ opacity: 0, y: 40 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: '-100px' }}
          transition={{ duration: 0.7 }}
        >
          <div className="inline-flex items-center gap-2 px-4 py-2 glass rounded-full mb-8">
            <Sparkles className="w-4 h-4 text-[#F59E0B]" />
            <span className="text-sm text-[var(--text-muted)]">Join 100K+ Users</span>
          </div>

          <h2 className="text-4xl md:text-6xl font-bold text-[var(--text-primary)] mb-6 leading-tight">
            Join The Future Of
            <br />
            <span className="gradient-text">Digital Finance</span>
          </h2>

          <p className="text-lg text-[var(--text-muted)] mb-10 max-w-2xl mx-auto">
            Start trading, learning, investing, and automating your financial future today.
          </p>

          <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
            <a
              href="#"
              className="w-full sm:w-auto px-8 py-4 text-lg font-semibold text-[var(--text-primary)] bg-gradient-to-r from-[#7C3AED] to-[#06B6D4] rounded-xl hover:opacity-90 transition-all duration-200 glow-purple flex items-center justify-center gap-2"
            >
              Create Free Account
              <ArrowRight className="w-5 h-5" />
            </a>
            <a
              href="#"
              className="w-full sm:w-auto px-8 py-4 text-lg font-semibold text-[var(--text-primary)] glass rounded-xl hover:bg-[var(--card-bg)]/50 transition-all duration-200 flex items-center justify-center gap-2"
            >
              <Calendar className="w-5 h-5" />
              Book Demo
            </a>
          </div>
        </motion.div>
      </div>
    </section>
  );
}
