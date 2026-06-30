'use client';

import { motion } from 'framer-motion';
import { Star, Quote } from 'lucide-react';

const testimonials = [
  {
    name: 'Adebayo K.',
    role: 'Crypto Trader',
    location: 'Lagos, Nigeria',
    image: 'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=100&h=100&fit=crop',
    quote: 'Coindistro transformed my trading. The AI bots alone have paid for my subscription 10x over. The signals are incredibly accurate.',
    profit: '+$12,450',
    rating: 5,
  },
  {
    name: 'Sarah M.',
    role: 'Investor',
    location: 'London, UK',
    image: 'https://images.unsplash.com/photo-1494790108377-be9c29b29330?w=100&h=100&fit=crop',
    quote: 'The Academy courses are world-class. I went from complete beginner to confident trader in just 3 months. Highly recommend!',
    profit: '+$8,320',
    rating: 5,
  },
  {
    name: 'Chen W.',
    role: 'Software Engineer',
    location: 'Shanghai, China',
    image: 'https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?w=100&h=100&fit=crop',
    quote: 'As a developer, I appreciate the technical excellence. The API is robust, documentation is clear, and the platform is rock solid.',
    profit: '+$15,780',
    rating: 5,
  },
];

export default function TestimonialsSection() {
  return (
    <section id="testimonials" className="relative py-24 overflow-hidden">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <motion.div
          initial={{ opacity: 0, y: 40 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: '-100px' }}
          transition={{ duration: 0.7 }}
          className="text-center mb-16"
        >
          <h2 className="text-3xl md:text-5xl font-bold text-[var(--text-primary)] mb-6">
            Trusted by <span className="gradient-text">Traders</span> Worldwide
          </h2>
          <p className="text-lg text-[var(--text-muted)] max-w-2xl mx-auto">
            See what our community members are saying about their experience.
          </p>
        </motion.div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          {testimonials.map((testimonial, i) => (
            <motion.div
              key={i}
              initial={{ opacity: 0, y: 30 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: i * 0.15 }}
              className="glass-card p-6 group hover:border-[#7C3AED]/40 transition-all duration-300"
            >
              {/* Quote icon */}
              <Quote className="w-8 h-8 text-[#7C3AED]/40 mb-4" />
              
              {/* Text */}
              <p className="text-[var(--text-muted)] text-sm leading-relaxed mb-6">
                "{testimonial.quote}"
              </p>

              {/* Profit badge */}
              <div className="inline-flex items-center gap-2 px-3 py-1.5 bg-[#10B981]/10 border border-[#10B981]/20 rounded-full mb-6">
                <Star className="w-4 h-4 text-[#10B981]" />
                <span className="text-sm font-semibold text-[#10B981]">{testimonial.profit} profit</span>
              </div>

              {/* Author */}
              <div className="flex items-center gap-4 pt-4 border-t border-[var(--card-border)]">
                <div className="w-12 h-12 rounded-full bg-gradient-to-br from-[#7C3AED] to-[#06B6D4] flex items-center justify-center text-[var(--text-primary)] font-bold">
                  {testimonial.name[0]}
                </div>
                <div>
                  <div className="font-semibold text-[var(--text-primary)]">{testimonial.name}</div>
                  <div className="text-xs text-[var(--text-muted)]">{testimonial.role} · {testimonial.location}</div>
                </div>
              </div>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}
