
import { useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Calendar, Clock, MapPin, Zap, Shield, TrendingUp } from "lucide-react";
import EventGrid from "@/components/EventGrid";
import SeatSelection from "@/components/SeatSelection";
import BookingFlow from "@/components/BookingFlow";
import { Event } from "@/services/api";

const Index = () => {
  const [selectedEvent, setSelectedEvent] = useState<Event | null>(null);
  const [selectedSeats, setSelectedSeats] = useState<string[]>([]);
  const [bookingStep, setBookingStep] = useState<'events' | 'seats' | 'checkout'>('events');

  const handleEventSelect = (event: Event) => {
    setSelectedEvent(event);
    setBookingStep('seats');
  };

  const handleSeatSelect = (seats: string[]) => {
    setSelectedSeats(seats);
    if (seats.length > 0) {
      setBookingStep('checkout');
    }
  };

  const handleBackToEvents = () => {
    setBookingStep('events');
    setSelectedEvent(null);
    setSelectedSeats([]);
  };

  const handleBackToSeats = () => {
    setBookingStep('seats');
    setSelectedSeats([]);
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric'
    });
  };

  const formatTime = (dateString: string) => {
    return new Date(dateString).toLocaleTimeString('en-US', {
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 to-blue-50">
      {/* Header */}
      <header className="booking-gradient text-white shadow-lg">
        <div className="container mx-auto px-4 py-6">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-3">
              <div className="p-2 bg-white/20 rounded-lg">
                <Zap className="h-8 w-8 text-white" />
              </div>
              <div>
                <h1 className="text-2xl font-bold">SeatSync Reserve</h1>
                <p className="text-blue-100">High-Performance Ticket Booking</p>
              </div>
            </div>
            <div className="flex items-center space-x-6 text-sm">
              <div className="flex items-center space-x-2">
                <Shield className="h-4 w-4" />
                <span>Secure</span>
              </div>
              <div className="flex items-center space-x-2">
                <TrendingUp className="h-4 w-4" />
                <span>Real-time</span>
              </div>
              <div className="flex items-center space-x-2">
                <Zap className="h-4 w-4" />
                <span>Instant</span>
              </div>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="container mx-auto px-4 py-8">
        {/* Breadcrumb */}
        <div className="flex items-center space-x-2 text-sm text-muted-foreground mb-6">
          <button 
            onClick={handleBackToEvents}
            className={`hover:text-primary transition-colors ${bookingStep === 'events' ? 'text-primary font-medium' : ''}`}
          >
            Events
          </button>
          {bookingStep !== 'events' && (
            <>
              <span>/</span>
              <button 
                onClick={handleBackToSeats}
                className={`hover:text-primary transition-colors ${bookingStep === 'seats' ? 'text-primary font-medium' : ''}`}
              >
                Seat Selection
              </button>
            </>
          )}
          {bookingStep === 'checkout' && (
            <>
              <span>/</span>
              <span className="text-primary font-medium">Checkout</span>
            </>
          )}
        </div>

        {/* Step Content */}
        {bookingStep === 'events' && (
          <div>
            <div className="text-center mb-8">
              <h2 className="text-3xl font-bold mb-4">Premium Events</h2>
              <p className="text-lg text-muted-foreground">
                Experience seamless booking with our high-performance system
              </p>
            </div>
            <EventGrid onEventSelect={handleEventSelect} />
          </div>
        )}

        {bookingStep === 'seats' && selectedEvent && (
          <div>
            <Card className="mb-6">
              <CardHeader>
                <div className="flex items-start justify-between">
                  <div>
                    <CardTitle className="text-2xl">{selectedEvent.name}</CardTitle>
                    <div className="text-lg mt-2 text-muted-foreground">
                      <div className="flex items-center space-x-4">
                        <div className="flex items-center space-x-1">
                          <Calendar className="h-4 w-4" />
                          <span>{formatDate(selectedEvent.start_time)}</span>
                        </div>
                        <div className="flex items-center space-x-1">
                          <Clock className="h-4 w-4" />
                          <span>{formatTime(selectedEvent.start_time)}</span>
                        </div>
                        <div className="flex items-center space-x-1">
                          <MapPin className="h-4 w-4" />
                          <span>{selectedEvent.venue}</span>
                        </div>
                      </div>
                    </div>
                  </div>
                  <Badge variant="secondary" className="text-lg px-3 py-1">
                    From ${selectedEvent.price}
                  </Badge>
                </div>
              </CardHeader>
            </Card>
            <SeatSelection 
              eventId={selectedEvent.id.toString()} 
              onSeatSelect={handleSeatSelect}
              selectedSeats={selectedSeats}
              eventPrice={selectedEvent.price}
            />
          </div>
        )}

        {bookingStep === 'checkout' && selectedEvent && (
          <BookingFlow 
            event={selectedEvent}
            selectedSeats={selectedSeats}
            onBack={handleBackToSeats}
          />
        )}
      </main>
    </div>
  );
};

export default Index;
