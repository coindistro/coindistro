'use client';

import { motion } from 'framer-motion';
import { Shield, Clock, Globe, Server, Headphones, Bot } from 'lucide-react';

const trustFeatures = [
  { icon: Shield, title: 'Bank-Level Security', desc: 'AES-256 encryption & cold storage' },
  { icon: Clock, title: 'Real-Time Execution', desc: 'Sub-millisecond order processing' },
  { icon: Globe, title: 'Global Access', desc: 'Trade from anywhere, 24/7/365' },
  { icon: Server, title: 'Institutional Infrastructure', desc: '99.99% uptime guarantee' },
  { icon: Headphones, title: '24/7 Support', desc: 'Live chat & phone support' },
  { icon: Bot, title: 'AI Automation', desc: 'Smart bots & copy trading' },
];

const stats = [
  { value: '$500M+', label: 'Projected Transaction Volume' },
  { value: '100K+', label: 'Future Users' },
  { value: '100+', label: 'Countries' },
  { value: '99.99%', label: 'Uptime SLA' },
];

export default function TrustSection() {
  return (
    <section id="trust" className="relative py-24 overflow-hidden">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        {/* Stats */}
        <motion.div
          initial={{ opacity: 0, y: 40 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: '-100px' }}
          transition={{ duration: 0.7 }}
          className="grid grid-cols-2 md:grid-cols-4 gap-6 mb-20"
        >
          {stats.map((stat, i) => (
            <div key={i} className="glass-card p-6 text-center group hover:border-[#7C3AED]/40 transition-colors duration-300">
              <div className="text-3xl md:text-4xl font-bold gradient-text mb-2">{stat.value}</div>
              <div className="text-sm text-[var(--text-muted)]">{stat.label}</div>
            </div>
          ))}
        </motion.div>

        {/* Trust Features */}
        <motion.div
          initial={{ opacity: 0, y: 40 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: '-100px' }}
          transition={{ duration: 0.7 }}
        >
          <div className="text-center mb-12">
            <h2 className="text-3xl md:text-4xl font-bold text-[var(--text-primary)] mb-4">
              Trusted by <span className="gradient-text">Thousands</span>
            </h2>
            <p className="text-[var(--text-muted)] max-w-2xl mx-auto">
              Built with security and reliability at its core, Coindistro provides institutional-grade infrastructure for everyone.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {trustFeatures.map((feature, i) => (
              <motion.div
                key={i}
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.5, delay: i * 0.1 }}
                className="glass-card p-6 group hover:border-[#7C3AED]/30 transition-all duration-300"
              >
                <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-[#7C3AED]/20 to-[#06B6D4]/20 flex items-center justify-center mb-4 group-hover:scale-110 transition-transform duration-300">
                  <feature.icon className="w-6 h-6 text-[#7C3AED]" />
                </div>
                <h3 className="text-lg font-semibold text-[var(--text-primary)] mb-2">{feature.title}</h3>
                <p className="text-sm text-[var(--text-muted)]">{feature.desc}</p>
              </motion.div>
            ))}
          </div>
        </motion.div>
      </div>
    </section>
  );
}
