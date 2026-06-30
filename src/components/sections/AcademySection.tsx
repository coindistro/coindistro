'use client';

import { motion } from 'framer-motion';
import { GraduationCap, BookOpen, Video, Award, Users, Star, Clock, ChevronRight } from 'lucide-react';

const courses = [
  {
    icon: BookOpen,
    title: 'Crypto Fundamentals',
    level: 'Beginner',
    duration: '4 weeks',
    students: '12.5K',
    rating: 4.9,
    progress: 0,
    modules: 8,
  },
  {
    icon: Video,
    title: 'Advanced Trading',
    level: 'Intermediate',
    duration: '6 weeks',
    students: '8.3K',
    rating: 4.8,
    progress: 0,
    modules: 12,
  },
  {
    icon: Award,
    title: 'Technical Analysis Mastery',
    level: 'Advanced',
    duration: '8 weeks',
    students: '5.1K',
    rating: 4.9,
    progress: 0,
    modules: 16,
  },
  {
    icon: BookOpen,
    title: 'Blockchain Development',
    level: 'Advanced',
    duration: '10 weeks',
    students: '3.2K',
    rating: 4.7,
    progress: 0,
    modules: 20,
  },
];

const stats = [
  { value: '50+', label: 'Courses' },
  { value: '25K+', label: 'Students' },
  { value: '4.8', label: 'Avg Rating' },
  { value: '95%', label: 'Completion' },
];

export default function AcademySection() {
  return (
    <section id="academy" className="relative py-24 overflow-hidden">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <motion.div
          initial={{ opacity: 0, y: 40 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: '-100px' }}
          transition={{ duration: 0.7 }}
          className="text-center mb-16"
        >
          <div className="inline-flex items-center gap-2 px-4 py-2 glass rounded-full mb-6">
            <GraduationCap className="w-4 h-4 text-[#F59E0B]" />
            <span className="text-sm text-[var(--text-muted)]">Coindistro Academy</span>
          </div>
          <h2 className="text-3xl md:text-5xl font-bold text-[var(--text-primary)] mb-6">
            Master <span className="gradient-text">Crypto Trading</span>
          </h2>
          <p className="text-lg text-[var(--text-muted)] max-w-2xl mx-auto">
            From beginner to pro, learn everything you need to succeed in the crypto markets.
          </p>
        </motion.div>

        {/* Stats */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
          className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-12"
        >
          {stats.map((stat, i) => (
            <div key={i} className="glass-card p-4 text-center">
              <div className="text-2xl font-bold gradient-text">{stat.value}</div>
              <div className="text-xs text-[var(--text-muted)] mt-1">{stat.label}</div>
            </div>
          ))}
        </motion.div>

        {/* Courses */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {courses.map((course, i) => (
            <motion.div
              key={i}
              initial={{ opacity: 0, y: 30 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: i * 0.1 }}
              className="glass-card p-6 group hover:border-[#F59E0B]/40 transition-all duration-300 cursor-pointer"
            >
              <div className="flex items-start gap-4">
                <div className="w-14 h-14 rounded-xl bg-gradient-to-br from-[#F59E0B] to-[#7C3AED] p-0.5 flex-shrink-0">
                  <div className="w-full h-full bg-[var(--card-bg)] rounded-xl flex items-center justify-center group-hover:bg-transparent transition-colors duration-300">
                    <course.icon className="w-7 h-7 text-[var(--text-primary)]" />
                  </div>
                </div>
                <div className="flex-1">
                  <div className="flex items-center gap-2 mb-2">
                    <span className={`px-2 py-0.5 text-xs rounded-full ${
                      course.level === 'Beginner' ? 'bg-[#10B981]/20 text-[#10B981]' :
                      course.level === 'Intermediate' ? 'bg-yellow-400/20 text-yellow-400' :
                      'bg-red-400/20 text-red-400'
                    }`}>
                      {course.level}
                    </span>
                    <span className="text-xs text-[var(--text-muted)] flex items-center gap-1">
                      <Clock className="w-3 h-3" />
                      {course.duration}
                    </span>
                  </div>
                  <h3 className="text-lg font-bold text-[var(--text-primary)] mb-2 group-hover:text-[#F59E0B] transition-colors duration-300">
                    {course.title}
                  </h3>
                  <div className="flex items-center gap-4 text-sm text-[var(--text-muted)] mb-4">
                    <span className="flex items-center gap-1">
                      <Users className="w-4 h-4" />
                      {course.students}
                    </span>
                    <span className="flex items-center gap-1">
                      <Star className="w-4 h-4 text-yellow-400" />
                      {course.rating}
                    </span>
                    <span className="flex items-center gap-1">
                      <BookOpen className="w-4 h-4" />
                      {course.modules} modules
                    </span>
                  </div>
                  <button className="flex items-center gap-1 text-sm text-[#7C3AED] hover:text-[#06B6D4] transition-colors duration-200">
                    Start Learning
                    <ChevronRight className="w-4 h-4" />
                  </button>
                </div>
              </div>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}
