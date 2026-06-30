'use client';

import { motion } from 'framer-motion';
import { Shield, Lock, Key, Eye, Fingerprint, Server, Globe } from 'lucide-react';

const securityFeatures = [
  {
    icon: Fingerprint,
    title: 'Multi-Factor Authentication',
    description: 'Secure your account with biometric, SMS, and hardware key authentication.',
  },
  {
    icon: Server,
    title: 'Cold Storage',
    description: '95% of assets stored in offline, air-gapped cold wallets.',
  },
  {
    icon: Lock,
    title: 'End-to-End Encryption',
    description: 'All data encrypted with AES-256 at rest and in transit.',
  },
  {
    icon: Eye,
    title: 'Fraud Detection AI',
    description: 'Real-time AI-powered monitoring detects and prevents suspicious activity.',
  },
  {
    icon: Shield,
    title: 'SOC 2 Compliance',
    description: 'Independently audited and certified for security controls.',
  },
  {
    icon: Globe,
    title: 'DDoS Protection',
    description: 'Enterprise-grade protection against distributed attacks.',
  },
];

export default function SecuritySection() {
  return (
    <section id="security" className="relative py-24 overflow-hidden">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <motion.div
          initial={{ opacity: 0, y: 40 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: '-100px' }}
          transition={{ duration: 0.7 }}
          className="text-center mb-16"
        >
          <div className="inline-flex items-center gap-2 px-4 py-2 glass rounded-full mb-6">
            <Shield className="w-4 h-4 text-[#10B981]" />
            <span className="text-sm text-[var(--text-muted)]">Bank-Level Security</span>
          </div>
          <h2 className="text-3xl md:text-5xl font-bold text-[var(--text-primary)] mb-6">
            Your Assets, <span className="gradient-text">Protected</span>
          </h2>
          <p className="text-lg text-[var(--text-muted)] max-w-2xl mx-auto">
            Military-grade security infrastructure to keep your funds and data safe.
          </p>
        </motion.div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {securityFeatures.map((feature, i) => (
            <motion.div
              key={i}
              initial={{ opacity: 0, y: 30 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: i * 0.1 }}
              className="glass-card p-6 group hover:border-[#10B981]/40 transition-all duration-300"
            >
              <div className="w-14 h-14 rounded-xl bg-gradient-to-br from-[#10B981]/20 to-[#06B6D4]/20 flex items-center justify-center mb-5 group-hover:scale-110 transition-transform duration-300">
                <feature.icon className="w-7 h-7 text-[#10B981]" />
              </div>
              <h3 className="text-lg font-bold text-[var(--text-primary)] mb-3 group-hover:text-[#10B981] transition-colors duration-300">
                {feature.title}
              </h3>
              <p className="text-sm text-[var(--text-muted)] leading-relaxed">{feature.description}</p>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}
