'use client';

import { motion } from 'framer-motion';
import { Bot, Cpu, Grid3X3, Signal, Users, Play, Settings } from 'lucide-react';

const bots = [
  {
    icon: Cpu,
    name: 'AI Scalper Bot',
    description: 'High-frequency AI-powered scalping with real-time market analysis.',
    monthlyReturn: '+24.8%',
    activeUsers: '2,847',
    status: 'Active',
    color: 'from-[#7C3AED] to-[#06B6D4]',
  },
  {
    icon: Grid3X3,
    name: 'Grid Trading Bot',
    description: 'Automated grid strategy for volatile market conditions.',
    monthlyReturn: '+18.3%',
    activeUsers: '1,923',
    status: 'Active',
    color: 'from-[#06B6D4] to-[#10B981]',
  },
  {
    icon: Bot,
    name: 'Smart DCA Bot',
    description: 'Dollar-cost averaging with intelligent entry timing.',
    monthlyReturn: '+12.5%',
    activeUsers: '3,156',
    status: 'Active',
    color: 'from-[#10B981] to-[#F59E0B]',
  },
  {
    icon: Signal,
    name: 'Signal Execution Bot',
    description: 'Auto-executes premium trading signals with precision.',
    monthlyReturn: '+21.2%',
    activeUsers: '1,445',
    status: 'Active',
    color: 'from-[#F59E0B] to-[#EF4444]',
  },
  {
    icon: Users,
    name: 'Copy Trading Bot',
    description: 'Mirror top traders automatically with risk management.',
    monthlyReturn: '+15.7%',
    activeUsers: '4,231',
    status: 'Active',
    color: 'from-[#EF4444] to-[#7C3AED]',
  },
];

export default function AIBotsSection() {
  return (
    <section id="bots" className="relative py-24 overflow-hidden">
      {/* Background glow */}
      <div className="absolute top-0 left-1/4 w-[600px] h-[600px] bg-[#7C3AED] rounded-full blur-[200px] opacity-10" />

      <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <motion.div
          initial={{ opacity: 0, y: 40 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: '-100px' }}
          transition={{ duration: 0.7 }}
          className="text-center mb-16"
        >
          <div className="inline-flex items-center gap-2 px-4 py-2 glass rounded-full mb-6">
            <Bot className="w-4 h-4 text-[#7C3AED]" />
            <span className="text-sm text-[var(--text-muted)]">AI-Powered Automation</span>
          </div>
          <h2 className="text-3xl md:text-5xl font-bold text-[var(--text-primary)] mb-6">
            AI Trading <span className="gradient-text">Bots</span>
          </h2>
          <p className="text-lg text-[var(--text-muted)] max-w-2xl mx-auto">
            Let advanced AI algorithms trade for you 24/7. Set up in minutes, profit while you sleep.
          </p>
        </motion.div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {bots.map((bot, i) => (
            <motion.div
              key={i}
              initial={{ opacity: 0, y: 30 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: i * 0.1 }}
              className="glass-card p-6 group hover:border-[#7C3AED]/40 transition-all duration-300"
            >
              {/* Bot Header */}
              <div className="flex items-start justify-between mb-6">
                <div className={`w-12 h-12 rounded-xl bg-gradient-to-br ${bot.color} p-0.5`}>
                  <div className="w-full h-full bg-[var(--card-bg)] rounded-xl flex items-center justify-center group-hover:bg-transparent transition-colors duration-300">
                    <bot.icon className="w-6 h-6 text-[var(--text-primary)]" />
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <span className="w-2 h-2 bg-[#10B981] rounded-full animate-pulse" />
                  <span className="text-xs text-[#10B981]">{bot.status}</span>
                </div>
              </div>

              {/* Bot Info */}
              <h3 className="text-xl font-bold text-[var(--text-primary)] mb-2 group-hover:text-[#7C3AED] transition-colors duration-300">
                {bot.name}
              </h3>
              <p className="text-sm text-[var(--text-muted)] mb-6 leading-relaxed">{bot.description}</p>

              {/* Stats */}
              <div className="grid grid-cols-2 gap-4 mb-6">
                <div className="glass p-3 rounded-lg">
                  <div className="text-xs text-[var(--text-muted)] mb-1">Monthly Return</div>
                  <div className="text-lg font-bold text-[#10B981]">{bot.monthlyReturn}</div>
                </div>
                <div className="glass p-3 rounded-lg">
                  <div className="text-xs text-[var(--text-muted)] mb-1">Active Users</div>
                  <div className="text-lg font-bold text-[var(--text-primary)]">{bot.activeUsers}</div>
                </div>
              </div>

              {/* Mini Chart */}
              <div className="flex items-end gap-1 h-12 mb-4">
                {Array.from({ length: 20 }).map((_, j) => {
                  const height = Math.random() * 80 + 20;
                  const isUp = Math.random() > 0.3;
                  return (
                    <div
                      key={j}
                      className={`flex-1 rounded-t-sm ${isUp ? 'bg-[#10B981]' : 'bg-red-400'}`}
                      style={{ height: `${height}%`, opacity: 0.6 + j * 0.02 }}
                    />
                  );
                })}
              </div>

              {/* Controls */}
              <div className="flex items-center gap-3">
                <button className="flex-1 py-2.5 text-sm font-medium text-[var(--text-primary)] bg-gradient-to-r from-[#7C3AED] to-[#06B6D4] rounded-lg hover:opacity-90 transition-all duration-200 flex items-center justify-center gap-2">
                  <Play className="w-4 h-4" />
                  Start Bot
                </button>
                <button className="p-2.5 text-[var(--text-muted)] glass rounded-lg hover:bg-[var(--card-bg)]/50 transition-colors duration-200">
                  <Settings className="w-4 h-4" />
                </button>
              </div>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}
