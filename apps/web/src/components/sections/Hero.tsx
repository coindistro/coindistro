'use client';

import { motion } from 'framer-motion';
import { ArrowRight, Play, TrendingUp, Wallet, Shield, Globe, Zap, Lock } from 'lucide-react';

const floatingIcons = [
  { Icon: TrendingUp, delay: 0, x: '10%', y: '20%' },
  { Icon: Wallet, delay: 2, x: '85%', y: '15%' },
  { Icon: Shield, delay: 4, x: '75%', y: '70%' },
  { Icon: Globe, delay: 1, x: '15%', y: '75%' },
  { Icon: Zap, delay: 3, x: '90%', y: '45%' },
  { Icon: Lock, delay: 5, x: '5%', y: '50%' },
];

export default function Hero() {
  return (
    <section className="relative min-h-screen flex items-center overflow-hidden pt-20">
      <div className="absolute inset-0 bg-[var(--background)]">
        <div className="absolute top-1/4 left-1/4 w-[600px] h-[600px] bg-[#7C3AED] rounded-full blur-[150px] opacity-20 animate-pulse" />
        <div className="absolute bottom-1/4 right-1/4 w-[400px] h-[400px] bg-[#06B6D4] rounded-full blur-[120px] opacity-15 animate-pulse" style={{ animationDelay: '2s' }} />
      </div>

      {floatingIcons.map(({ Icon, delay, x, y }, i) => (
        <motion.div
          key={i}
          className="absolute hidden lg:block"
          style={{ left: x, top: y }}
          animate={{ y: [0, -20, 0], opacity: [0.3, 0.6, 0.3] }}
          transition={{ duration: 4, delay, repeat: Infinity, ease: 'easeInOut' }}
        >
          <div className="w-12 h-12 glass rounded-xl flex items-center justify-center">
            <Icon className="w-6 h-6 text-[#7C3AED]" />
          </div>
        </motion.div>
      ))}

      <div className="relative z-10 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-20">
        <div className="text-center max-w-4xl mx-auto">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5 }}
            className="inline-flex items-center gap-2 px-4 py-2 glass rounded-full mb-8"
          >
            <span className="w-2 h-2 bg-[#10B981] rounded-full animate-pulse" />
            <span className="text-sm text-[var(--text-muted)]">Africa&apos;s Next-Gen Crypto Ecosystem</span>
          </motion.div>

          <motion.h1
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.7, delay: 0.1 }}
            className="text-5xl sm:text-6xl lg:text-7xl font-bold leading-tight mb-6"
          >
            <span className="text-[var(--text-primary)]">One Platform.</span>
            <br />
            <span className="gradient-text">Everything Crypto.</span>
          </motion.h1>

          <motion.p
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.7, delay: 0.2 }}
            className="text-lg sm:text-xl text-[var(--text-muted)] mb-10 max-w-2xl mx-auto leading-relaxed"
          >
            Trade, learn, automate, invest, and spend digital assets through Africa&apos;s next-generation crypto financial ecosystem.
          </motion.p>

          <motion.div
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.7, delay: 0.3 }}
            className="flex flex-col sm:flex-row items-center justify-center gap-4 mb-16"
          >
            <a href="/register" className="px-8 py-4 text-base font-semibold text-[var(--text-primary)] bg-gradient-to-r from-[#7C3AED] to-[#06B6D4] rounded-xl hover:opacity-90 transition-all duration-200 glow-purple flex items-center gap-2">
              Get Started <ArrowRight className="w-5 h-5" />
            </a>
            <a href="#ecosystem" className="px-8 py-4 text-base font-semibold text-[var(--text-primary)] glass rounded-xl hover:bg-[var(--card-bg)]/50 transition-all duration-200 flex items-center gap-2">
              <Play className="w-5 h-5" /> Explore Ecosystem
            </a>
          </motion.div>

          <motion.div
            initial={{ opacity: 0, y: 40 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.7, delay: 0.5 }}
            className="grid grid-cols-2 md:grid-cols-4 gap-4 max-w-3xl mx-auto"
          >
            {['$500M+', '100K+', '100+', '99.99%'].map((val, i) => (
              <div key={i} className="glass-card p-4 text-center">
                <div className="text-2xl font-bold gradient-text">{val}</div>
                <div className="text-xs text-[var(--text-muted)] mt-1">{['Projected Volume','Future Users','Countries','Uptime'][i]}</div>
              </div>
            ))}
          </motion.div>
        </div>
      </div>
    </section>
  );
}
