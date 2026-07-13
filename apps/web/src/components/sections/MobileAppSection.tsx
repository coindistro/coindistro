'use client';

import { motion } from 'framer-motion';
import { Smartphone, Home, BarChart3, ArrowLeftRight, Zap, Bot, GraduationCap, Wallet, User } from 'lucide-react';

const appTabs = [
  { icon: Home, label: 'Home', active: true },
  { icon: BarChart3, label: 'Markets', active: false },
  { icon: ArrowLeftRight, label: 'Trade', active: false },
  { icon: Zap, label: 'Signals', active: false },
  { icon: Bot, label: 'Bots', active: false },
  { icon: GraduationCap, label: 'Academy', active: false },
  { icon: Wallet, label: 'Wallet', active: false },
  { icon: User, label: 'Profile', active: false },
];

const portfolioData = [
  { name: 'BTC', value: 45.2, color: '#7C3AED' },
  { name: 'ETH', value: 28.5, color: '#06B6D4' },
  { name: 'SOL', value: 15.8, color: '#10B981' },
  { name: 'Others', value: 10.5, color: '#F59E0B' },
];

export default function MobileAppSection() {
  return (
    <section id="mobile" className="relative py-24 overflow-hidden">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <motion.div
          initial={{ opacity: 0, y: 40 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: '-100px' }}
          transition={{ duration: 0.7 }}
          className="text-center mb-16"
        >
          <div className="inline-flex items-center gap-2 px-4 py-2 glass rounded-full mb-6">
            <Smartphone className="w-4 h-4 text-[#7C3AED]" />
            <span className="text-sm text-[var(--text-muted)]">Mobile App</span>
          </div>
          <h2 className="text-3xl md:text-5xl font-bold text-[var(--text-primary)] mb-6">
            Crypto In Your <span className="gradient-text">Pocket</span>
          </h2>
          <p className="text-lg text-[var(--text-muted)] max-w-2xl mx-auto">
            Trade, learn, and manage your portfolio anywhere with our powerful mobile app.
          </p>
        </motion.div>

        <div className="flex flex-col lg:flex-row items-center justify-center gap-12">
          {/* awesomePhone Mockup */}
          <motion.div
            initial={{ opacity: 0, x: -40 }}
            whileInView={{ opacity: 1, x: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.7 }}
            className="relative"
          >
            <div className="w-[320px] h-[640px] bg-[var(--card-bg)] rounded-[40px] border-4 border-[var(--card-border)] p-2 relative overflow-hidden">
              {/* Screen bezel */}
              <div className="absolute top-4 left-1/2 -translate-x-1/2 w-24 h-6 bg-[var(--card-bg)] rounded-full z-10" />
              
              {/* Screen content */}
              <div className="w-full h-full bg-[var(--card-bg)] rounded-[32px] overflow-hidden flex flex-col">
                {/* Status bar */}
                <div className="flex items-center justify-between px-6 pt-2 pb-1">
                  <span className="text-xs text-[var(--text-muted)]">9:41</span>
                  <div className="flex items-center gap-1">
                    <div className="w-4 h-4 rounded-full bg-white/20" />
                    <div className="w-4 h-4 rounded-full bg-white/20" />
                  </div>
                </div>

                {/* App Header */}
                <div className="px-4 py-3">
                  <div className="flex items-center justify-between mb-2">
                    <div>
                      <p className="text-xs text-[var(--text-muted)]">Total Balance</p>
                      <p className="text-2xl font-bold text-[var(--text-primary)]">$24,562.80</p>
                      <p className="text-xs text-[#10B981]">+$2,720 (+12.4%)</p>
                    </div>
                    <div className="w-10 h-10 rounded-full bg-gradient-to-br from-[#7 displacedEEED] to-[#06B6D4] flex items-center justify-center">
                      <Wallet className="w-5 h-5 text-[var(--text-primary)]" />
                    </div>
                  </div>
                </div>

                {/* Chart area */}
                <div className="px-4 mb-4">
                  <div className="glass rounded-xl p-4">
                    <p className="text-xs text-[var(--text-muted)] mb-2">Portfolio</p>
                    <div className="flex items-end gap-2 h-24">
                      {portfolioData.map((item, i) => (
                        <div key={i} className="flex-1 flex flex-col items-center gap-1">
                          <div
                            className="w-full rounded-t-lg transition-all duration-500"
                            style={{
                              height: `${item.value * 2}px`,
                              backgroundColor: item.color,
                              opacity: 0.8,
                            }}
                          />
                          <span className="text-[10px] text-[var(--text-muted)]">{item.name}</span>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>

                {/* Quick actions */}
                <div className="px-4 mb-4">
                  <div className="grid grid-cols-4 gap-3">
                    {['Buy', 'Sell', 'Send', 'Receive'].map((action, i) => (
                      <div key={i} className="glass rounded-xl p-3 text-center">
                        <div className="w-8 h-8 rounded-full bg-gradient-to-br from-[#7C3AED] to-[#06B6D4] flex items-center justify-center mx-auto mb-1">
                          <ArrowLeftRight className="w-4 h-4 text-[var(--text-primary)]" />
                        </div>
                        <span className="text-[10px] text-[var(--text-muted)]">{action}</span>
                      </div>
                    ))}
                  </div>
                </div>

                {/* Recent coins */}
                <div className="px-4 flex-1">
                  <p className="text-xs text-[var(--text-muted)] mb-2">Watchlist</p>
                  <div className="space-y-2">
                    {[
                      { name: 'BTC', price: '97,245', change: '+2.4%' },
                      { name: 'ETH', price: '3,842', change: '+1.8%' },
                      { name: 'SOL', price: '198.50', change: '-0.5%' },
                    ].map((coin, i) => (
                      <div key={i} className="glass rounded-lg p-3 flex items-center justify-between">
                        <div className="flex items-center gap-2">
                          <div className="w-8 h-8 rounded-full bg-gradient-to-br from-[#7C3AED] to-[#06B6D4] flex items-center justify-center text-xs font-bold text-[var(--text-primary)]">
                            {coin.name[0]}
                          </div>
                          <span className="text-sm text-[var(--text-primary)] font-medium">{coin.name}</span>
                        </div>
                        <div className="text-right">
                          <p className="text-sm text-[var(--text-primary)]">${coin.price}</p>
                          <p className={coin.change.startsWith('+') ? 'text-xs text-[#10B981]' : 'text-xs text-red-400'}>
                            {coin.change}
                          </p>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>

                {/* Bottom tabs */}
                <div className="glass-t p-2 pb-4">
                  <div className="grid grid-cols-8 gap-1">
                    {appTabs.slice(0, 8).map((tab, i) => (
                      <div key={i} className="flex flex-col items-center gap-1 py-1">
                        <tab.icon className={`w-4 h-4 ${tab.active ? 'text-[#7C3AED]' : 'text-[var(--text-muted)]'}`} />
                        <span className={`text-[8px] ${tab.active ? 'text-[#7C3AED]' : 'text-[var(--text-muted)]'}`}>{tab.label}</span>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            </div>

            {/* Glow effect */}
            <div className="absolute -inset-4 bg-gradient-to-r from-[#7C3AED]/20 to-[#06B6D4]/20 rounded-[50px] blur-2xl -z-10" />
          </motion.div>

          {/* Features */}
          <motion.div
            initial={{ opacity: 0, x: 40 }}
            whileInView={{ opacity: 1, x: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.7, delay: 0.2 }}
            className="max-w-md"
          >
            <h3 className="text-2xl font-bold text-[var(--text-primary)] mb-6">
              Everything You Need, <span className="gradient-text">Anywhere</span>
            </h3>
            <div className="space-y-4">
              {[
                { icon: BarChart3, title: 'Real-Time Markets', desc: 'Live prices, charts, and market data at your fingertips.' },
                { icon: ArrowLeftRight, title: 'Instant Trading', desc: 'Buy, sell, and swap crypto in seconds with deep liquidity.' },
                { icon: Zap, title: 'Trading Signals', desc: 'Get professional signals delivered straight to your phone.' },
                { icon: Bot, title: 'AI Bots', desc: 'Deploy and monitor trading bots from anywhere.' },
                { icon: Wallet, title: 'Multi-Currency Wallet', desc: 'Securely store and manage all your digital assets.' },
              ].map((feature, i) => (
                <motion.div
                  key={i}
                  initial={{ opacity: 0, y: 20 }}
                  whileInView={{ opacity: 1, y: 0 }}
                  viewport={{ once: true }}
                  transition={{ duration: 0.5, delay: 0.3 + i * 0.1 }}
                  className="flex items-start gap-4 p-4 glass rounded-xl hover:border-[#7C3AED]/30 transition-colors duration-300"
                >
                  <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-[#7C3AED]/20 to-[#06B6D4]/20 flex items-center justify-center flex-shrink-0">
                    <feature.icon className="w-5 h-5 text-[#7C3AED]" />
                  </div>
                  <div>
                    <h4 className="text-sm font-semibold text-[var(--text-primary)] mb-1">{feature.title}</h4>
                    <p className="text-xs text-[var(--text-muted)] leading-relaxed">{feature.desc}</p>
                  </div>
                </motion.div>
              ))}
            </div>
          </motion.div>
        </div>
      </div>
    </section>
  );
}
