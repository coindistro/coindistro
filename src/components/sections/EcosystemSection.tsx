'use client';

import { motion } from 'framer-motion';
import { 
  ArrowRightLeft, CreditCard, Landmark, Gift, 
  GraduationCap, BarChart3, Bot, PiggyBank 
} from 'lucide-react';

const ecosystemProducts = [
  {
    icon: ArrowRightLeft,
    title: 'Coindistro Exchange',
    description: 'Spot trading, futures, and P2P marketplace with deep liquidity.',
    features: ['Spot Trading', 'Futures', 'P2P Marketplace'],
    color: 'from-[#7C3AED] to-[#06B6D4]',
  },
  {
    icon: CreditCard,
    title: 'Coindistro Pay',
    description: 'Merchant payments, payment gateway, and invoicing solutions.',
    features: ['Merchant Payments', 'Payment Gateway', 'Invoicing'],
    color: 'from-[#06B6D4] to-[#10B981]',
  },
  {
    icon: Landmark,
    title: 'Coindistro Bank',
    description: 'Digital banking, virtual accounts, and remittance services.',
    features: ['Digital Banking', 'Virtual Accounts', 'Remittance'],
    color: 'from-[#10B981] to-[#7C3AED]',
  },
  {
    icon: Gift,
    title: 'Coindistro Gift Cards',
    description: 'Buy and sell gift cards with instant settlement.',
    features: ['Buy & Sell', 'Instant Settlement', 'Global Brands'],
    color: 'from-[#7C3AED] to-[#EF4444]',
  },
  {
    icon: GraduationCap,
    title: 'Coindistro Academy',
    description: 'Crypto education, trading courses, and certifications.',
    features: ['Crypto Education', 'Trading Courses', 'Certifications'],
    color: 'from-[#F59E0B] to-[#7C3AED]',
  },
  {
    icon: BarChart3,
    title: 'Coindistro Signals',
    description: 'Market analysis, trading signals, and technical insights.',
    features: ['Market Analysis', 'Trading Signals', 'Technical Insights'],
    color: 'from-[#06B6D4] to-[#F59E0B]',
  },
  {
    icon: Bot,
    title: 'Coindistro Trading Bots',
    description: 'Grid bots, DCA bots, AI bots, and copy trading.',
    features: ['Grid Bots', 'DCA Bots', 'AI Bots', 'Copy Trading'],
    color: 'from-[#EF4444] to-[#7C3AED]',
  },
  {
    icon: PiggyBank,
    title: 'Coindistro Invest',
    description: 'Staking, yield products, and portfolio management.',
    features: ['Staking', 'Yield Products', 'Portfolio Management'],
    color: 'from-[#10B981] to-[#06B6D4]',
  },
];

export default function EcosystemSection() {
  return (
    <section id="ecosystem" className="relative py-24 overflow-hidden">
      {/* Background glow */}
      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[800px] h-[800px] bg-[#7C3AED] rounded-full blur-[200px] opacity-10" />

      <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <motion.div
          initial={{ opacity: 0, y: 40 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: '-100px' }}
          transition={{ duration: 0.7 }}
          className="text-center mb-16"
        >
          <h2 className="text-3xl md:text-5xl font-bold text-[var(--text-primary)] mb-6">
            The <span className="gradient-text">Coindistro</span> Ecosystem
          </h2>
          <p className="text-lg text-[var(--text-muted)] max-w-2xl mx-auto">
            Everything you need to trade, learn, invest, and spend in the crypto economy — all in one powerful platform.
          </p>
        </motion.div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          {ecosystemProducts.map((product, i) => (
            <motion.div
              key={i}
              initial={{ opacity: 0, y: 30 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: i * 0.08 }}
              className="glass-card p-6 group hover:border-[#7C3AED]/40 transition-all duration-300 cursor-pointer"
            >
              <div className={`w-14 h-14 rounded-xl bg-gradient-to-br ${product.color} p-0.5 mb-5`}>
                <div className="w-full h-full bg-[var(--card-bg)] rounded-xl flex items-center justify-center group-hover:bg-transparent transition-colors duration-300">
                  <product.icon className="w-7 h-7 text-[var(--text-primary)]" />
                </div>
              </div>
              
              <h3 className="text-xl font-bold text-[var(--text-primary)] mb-3 group-hover:text-[#7C3AED] transition-colors duration-300">
                {product.title}
              </h3>
              <p className="text-sm text-[var(--text-muted)] mb-4 leading-relaxed">{product.description}</p>
              
              <div className="flex flex-wrap gap-2">
                {product.features.map((feature, j) => (
                  <span key={j} className="px-3 py-1 text-xs glass rounded-full text-[var(--text-muted)]">
                    {feature}
                  </span>
                ))}
              </div>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}
