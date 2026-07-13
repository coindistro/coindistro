'use client';

import { motion } from 'framer-motion';
import { Shield, Target, AlertTriangle } from 'lucide-react';

const tradingSignals = [
  {
    pair: 'BTCUSDT',
    entry: 115000,
    stopLoss: 113500,
    takeProfit: 117000,
    riskLevel: 'Medium',
    signal: 'BUY',
    confidence: 87,
  },
  {
    pair: 'ETHUSDT',
    entry: 5800,
    stopLoss: 5700,
    takeProfit: 6100,
    riskLevel: 'Low',
    signal: 'BUY',
    confidence: 92,
  },
  {
    pair: 'SOLUSDT',
    entry: 210,
    stopLoss: 195,
    takeProfit: 240,
    riskLevel: 'High',
    signal: 'BUY',
    confidence: 74,
  },
  {
    pair: 'XRPUSDT',
    entry: 2.85,
    stopLoss: 2.65,
    takeProfit: 3.20,
    riskLevel: 'Medium',
    signal: 'BUY',
    confidence: 81,
  },
];

export default function TradingSignalsSection() {
  return (
    <section id="signals" className="relative py-24 overflow-hidden">
      {/* Background glow */}
      <div className="absolute top-1/2 right-0 w-[500px] h-[500px] bg-[#06B6D4] rounded-full blur-[200px] opacity-10" />

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <motion.div
          initial={{ opacity: 0, y: 40 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: '-100px' }}
          transition={{ duration: 0.7 }}
          className="text-center mb-16"
        >
          <div className="inline-flex items-center gap-2 px-4 py-2 glass rounded-full mb-6">
            <Target className="w-4 h-4 text-[#06B6D4]" />
            <span className="text-sm text-[var(--text-muted)]">Professional Trading Signals</span>
          </div>
          <h2 className="text-3xl md:text-5xl font-bold text-[var(--text-primary)] mb-6">
            Expert <span className="gradient-text">Trading Signals</span>
          </h2>
          <p className="text-lg text-[var(--text-muted)] max-w-2xl mx-auto">
            Confident, data-driven trading signals with clear entry, stop loss, and take profit levels.
          </p>
        </motion.div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {tradingSignals.map((signal, i) => (
            <motion.div
              key={i}
              initial={{ opacity: 0, y: 30 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: i * 0.1 }}
              className="glass-card p-6 group hover:border-[#7C3AED]/40 transition-all duration-300"
            >
              {/* Header */}
              <div className="flex items-center justify-between mb-6">
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-[#7C3AED] to-[#06B6D4] flex items-center justify-center">
                    <span className="text-sm font-bold text-[var(--text-primary)]">{signal.pair.slice(0, 3)}</span>
                  </div>
                  <div>
                    <h3 className="text-lg font-bold text-[var(--text-primary)]">{signal.pair}</h3>
                    <div className="flex items-center gap-2 text-xs text-[var(--text-muted)]">
                      <span className={`px-2 py-0.5 rounded-full ${signal.signal === 'BUY' ? 'bg-[#10B981]/20 text-[#10B981]' : 'bg-red-400/20 text-red-400'}`}>
                        {signal.signal}
                      </span>
                      <span className={`flex items-center gap-1 ${signal.riskLevel === 'Low' ? 'text-[#10B981]' : signal.riskLevel === 'Medium' ? 'text-yellow-400' : 'text-red-400'}`}>
                        <Shield className="w-3 h-3" />
                        {signal.riskLevel} Risk
                      </span>
                    </div>
                  </div>
                </div>
                <div className="text-right">
                  <div className="text-2xl font-bold gradient-text">{signal.confidence}%</div>
                  <div className="text-xs text-[var(--text-muted)]">Confidence</div>
                </div>
              </div>

              {/* Levels */}
              <div className="grid grid-cols-3 gap-4 mb-4">
                <div className="glass p-3 rounded-lg">
                  <div className="text-xs text-[var(--text-muted)] mb-1">Entry</div>
                  <div className="text-lg font-bold text-[var(--text-primary)]">${signal.entry.toLocaleString()}</div>
                </div>
                <div className="glass p-3 rounded-lg">
                  <div className="text-xs text-[var(--text-muted)] mb-1">Stop Loss</div>
                  <div className="text-lg font-bold text-red-400">${signal.stopLoss.toLocaleString()}</div>
                </div>
                <div className="glass p-3 rounded-lg">
                  <div className="text-xs text-[var(--text-muted)] mb-1">Take Profit</div>
                  <div className="text-lg font-bold text-[#10B981]">${signal.takeProfit.toLocaleString()}</div>
                </div>
              </div>

              {/* Progress bar */}
              <div className="w-full bg-white/5 rounded-full h-2 overflow-hidden">
                <motion.div
                  initial={{ width: 0 }}
                  whileInView={{ width: `${signal.confidence}%` }}
                  viewport={{ once: true }}
                  transition={{ duration: 1, delay: 0.5 }}
                  className="h-full bg-gradient-to-r from-[#7C3AED] to-[#06B6D4] rounded-full"
                />
              </div>
            </motion.div>
          ))}
        </div>

        {/* Disclaimer */}
        <motion.div
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5, delay: 0.8 }}
          className="flex items-start gap-3 mt-8 p-4 glass rounded-xl"
        >
          <AlertTriangle className="w-5 h-5 text-yellow-400 flex-shrink-0 mt-0.5" />
          <p className="text-sm text-[var(--text-muted)]">
            Trading signals are for educational purposes only. Past performance does not guarantee future results. 
            Always do your own research and never invest more than you can afford to lose.
          </p>
        </motion.div>
      </div>
    </section>
  );
}
