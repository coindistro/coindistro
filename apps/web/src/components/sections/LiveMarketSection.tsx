'use client';

import { motion } from 'framer-motion';
import { TrendingUp, TrendingDown } from 'lucide-react';

const marketData = [
  { pair: 'BTC/USDT', price: 97245.80, change: 2.41, volume: '32.4B', isPositive: true },
  { pair: 'ETH/USDT', price: 3842.15, change: 1.85, volume: '18.7B', isPositive: true },
  { pair: 'SOL/USDT', price: 198.50, change: -0.53, volume: '4.2B', isPositive: false },
  { pair: 'XRP/USDT', price: 2.84, change: 5.12, volume: '3.8B', isPositive: true },
  { pair: 'BNB/USDT', price: 725.40, change: -1.23, volume: '1.9B', isPositive: false },
];

export default function LiveMarketSection() {
  return (
    <section id="market" className="relative py-24 overflow-hidden">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <motion.div
          initial={{ opacity: 0, y: 40 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: '-100px' }}
          transition={{ duration: 0.7 }}
          className="text-center mb-16"
        >
          <h2 className="text-3xl md:text-5xl font-bold text-[var(--text-primary)] mb-6">
            Live <span className="gradient-text">Markets</span>
          </h2>
          <p className="text-lg text-[var(--text-muted)] max-w-2xl mx-auto">
            Real-time prices and market data for the top cryptocurrencies.
          </p>
        </motion.div>

        {/* Market Table */}
        <motion.div
          initial={{ opacity: 0, y: 40 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.7 }}
          className="glass-card overflow-hidden"
        >
          {/* Header */}
          <div className="grid grid-cols-5 gap-4 p-4 border-b border-[var(--card-border)] text-xs font-medium text-[var(--text-muted)] uppercase tracking-wider">
            <div>Pair</div>
            <div className="text-right">Price</div>
            <div className="text-right">24h Change</div>
            <div className="text-right">Volume</div>
            <div className="text-right">Chart</div>
          </div>

          {/* Rows */}
          {marketData.map((item, i) => (
            <motion.div
              key={i}
              initial={{ opacity: 0, x: -20 }}
              whileInView={{ opacity: 1, x: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.4, delay: i * 0.1 }}
              className="grid grid-cols-5 gap-4 p-4 border-b border-[var(--card-border)] hover:bg-white/5 transition-colors duration-200 items-center"
            >
              <div className="flex items-center gap-3">
                <div className="w-8 h-8 rounded-full bg-gradient-to-br from-[#7C3AED] to-[#06B6D4] flex items-center justify-center text-xs font-bold text-[var(--text-primary)]">
                  {item.pair.split('/')[0][0]}
                </div>
                <span className="font-semibold text-[var(--text-primary)]">{item.pair}</span>
              </div>
              <div className="text-right font-mono text-[var(--text-primary)]">${item.price.toLocaleString('en-US', { minimumFractionDigits: item.price < 10 ? 2 : 2 })}</div>
              <div className={`text-right flex items-center justify-end gap-1 ${item.isPositive ? 'text-[#10B981]' : 'text-red-400'}`}>
                {item.isPositive ? <TrendingUp className="w-4 h-4" /> : <TrendingDown className="w-4 h-4" />}
                <span>{item.isPositive ? '+' : ''}{item.change}%</span>
              </div>
              <div className="text-right text-[var(--text-muted)]">{item.volume}</div>
              <div className="flex items-center justify-end gap-1 h-8">
                {Array.from({ length: 12 }).map((_, j) => {
                  const height = Math.random() * 60 + 20;
                  const isUp = Math.random() > 0.5;
                  return (
                    <div
                      key={j}
                      className={`w-1 rounded-full ${isUp ? 'bg-[#10B981]' : 'bg-red-400'}`}
                      style={{ height: `${height}%`, opacity: 0.7 + j * 0.02 }}
                    />
                  );
                })}
              </div>
            </motion.div>
          ))}
        </motion.div>
      </div>
    </section>
  );
}
