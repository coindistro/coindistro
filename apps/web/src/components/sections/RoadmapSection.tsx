'use client';

import { motion } from 'framer-motion';
import { Check, Clock } from 'lucide-react';

const phases = [
  {
    phase: 'Phase 1',
    title: 'Foundation',
    status: 'completed',
    items: ['Exchange Launch', 'Payment Gateway', 'Gift Cards', 'Academy'],
  },
  {
    phase: 'Phase 2',
    title: 'Growth',
    status: 'completed',
    items: ['Trading Signals', 'Market Analysis', 'Community Features'],
  },
  {
    phase: 'Phase 3',
    title: 'Automation',
    status: 'in-progress',
    items: ['Trading Bots', 'Copy Trading', 'Portfolio Management'],
  },
  {
    phase: 'Phase 4',
    title: 'Global Scale',
    status: 'upcoming',
    items: ['Digital Banking', 'Investments', 'Merchant APIs', 'Global Expansion'],
  },
];

export default function RoadmapSection() {
  return (
    <section id="roadmap" className="relative py-24 overflow-hidden">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <motion.div
          initial={{ opacity: 0, y: 40 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: '-100px' }}
          transition={{ duration: 0.7 }}
          className="text-center mb-16"
        >
          <h2 className="text-3xl md:text-5xl font-bold text-[var(--text-primary)] mb-6">
            Our <span className="gradient-text">Roadmap</span>
          </h2>
          <p className="text-lg text-[var(--text-muted)] max-w-2xl mx-auto">
            Building the future of crypto finance, one milestone at a time.
          </p>
        </motion.div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          {phases.map((phase, i) => (
            <motion.div
              key={i}
              initial={{ opacity: 0, y: 30 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: i * 0.15 }}
              className={`glass-card p-6 ${
                phase.status === 'completed' ? 'border-[#10B981]/30' :
                phase.status === 'in-progress' ? 'border-[#F59E0B]/30' :
                'border-[var(--card-border)]'
              } transition-all duration-300`}
            >
              {/* Phase Header */}
              <div className="flex items-center justify-between mb-4">
                <span className={`text-sm font-medium ${
                  phase.status === 'completed' ? 'text-[#10B981]' :
                  phase.status === 'in-progress' ? 'text-[#F59E0B]' :
                  'text-[var(--text-muted)]'
                }`}>
                  {phase.phase}
                </span>
                <span className={`px-2 py-0.5 text-xs rounded-full ${
                  phase.status === 'completed' ? 'bg-[#10B981]/20 text-[#10B981]' :
                  phase.status === 'in-progress' ? 'bg-[#F59E0B]/20 text-[#F59E0B]' :
                  'bg-white/10 text-[var(--text-muted)]'
                }`}>
                  {phase.status === 'completed' ? 'Done' : phase.status === 'in-progress' ? 'In Progress' : 'Upcoming'}
                </span>
              </div>

              <h3 className="text-xl font-bold text-[var(--text-primary)] mb-4">{phase.title}</h3>

              <ul className="space-y-3">
                {phase.items.map((item, j) => (
                  <li key={j} className="flex items-center gap-3 text-sm text-[var(--text-muted)]">
                    <div className={`w-5 h-5 rounded-full flex items-center justify-center flex-shrink-0 ${
                      phase.status === 'completed' ? 'bg-[#10B981]/20' :
                      phase.status === 'in-progress' ? 'bg-[#F59E0B]/20' :
                      'bg-white/10'
                    }`}>
                      {phase.status === 'completed' ? (
                        <Check className="w-3 h-3 text-[#10B981]" />
                      ) : phase.status === 'in-progress' ? (
                        <Clock className="w-3 h-3 text-[#F59E0B]" />
                      ) : (
                        <div className="w-1.5 h-1.5 rounded-full bg-[#B8C0CC]" />
                      )}
                    </div>
                    {item}
                  </li>
                ))}
              </ul>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}
