import Navbar from '@/components/Navbar';
import Hero from '@/components/sections/Hero';
import TrustSection from '@/components/sections/TrustSection';
import EcosystemSection from '@/components/sections/EcosystemSection';
import LiveMarketSection from '@/components/sections/LiveMarketSection';
import TradingSignalsSection from '@/components/sections/TradingSignalsSection';
import AIBotsSection from '@/components/sections/AIBotsSection';
import AcademySection from '@/components/sections/AcademySection';
import MobileAppSection from '@/components/sections/MobileAppSection';
import SecuritySection from '@/components/sections/SecuritySection';
import TestimonialsSection from '@/components/sections/TestimonialsSection';
import RoadmapSection from '@/components/sections/RoadmapSection';
import CTASection from '@/components/sections/CTASection';
import Footer from '@/components/Footer';

export default function Home() {
  return (
    <main className="min-h-screen bg-[var(--background)] text-[var(--text-primary)] overflow-x-hidden scroll-smooth">
      <Navbar />
      <Hero />
      <TrustSection />
      <EcosystemSection />
      <LiveMarketSection />
      <TradingSignalsSection />
      <AIBotsSection />
      <AcademySection />
      <MobileAppSection />
      <SecuritySection />
      <TestimonialsSection />
      <RoadmapSection />
      <CTASection />
      <Footer />
    </main>
  );
}
