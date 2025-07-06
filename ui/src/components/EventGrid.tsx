import { useEffect, useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Calendar, Clock, MapPin, Users, Star } from "lucide-react";
import { useQuery } from "@tanstack/react-query";
import { apiService, Event } from "@/services/api";
import { toast } from "sonner";
import { getConfig } from "@/lib/config";

interface EventGridProps {
  onEventSelect: (event: Event) => void;
}

// Event categories and their associated keywords
const EVENT_CATEGORIES = {
  'Music': ['music', 'concert', 'jazz', 'rock', 'classical', 'band', 'orchestra', 'singer'],
  'Conference': ['tech', 'conference', 'summit', 'convention', 'meetup', 'workshop', 'seminar'],
  'Theater': ['theater', 'musical', 'play', 'drama', 'comedy', 'show', 'performance'],
  'Food': ['food', 'wine', 'tasting', 'culinary', 'dining', 'restaurant', 'chef'],
  'Sports': ['sport', 'championship', 'game', 'match', 'tournament', 'league', 'cup'],
  'Art': ['art', 'gallery', 'exhibition', 'museum', 'painting', 'sculpture', 'artist'],
  'Gaming': ['gaming', 'esports', 'tournament', 'competition', 'video game', 'console'],
  'Festival': ['festival', 'celebration', 'carnival', 'fair', 'parade']
};

const EVENT_EMOJIS = {
  'Music': '🎵',
  'Conference': '💻',
  'Theater': '🎭',
  'Food': '🍷',
  'Sports': '🏆',
  'Art': '🎨',
  'Gaming': '🎮',
  'Festival': '🎪',
  'Event': '🎪'
};

const EventGrid = ({ onEventSelect }: EventGridProps) => {
  const config = getConfig();
  
  const { data: eventsResponse, isLoading, error } = useQuery({
    queryKey: ['events'],
    queryFn: () => apiService.getEvents(1, 20), // Could be made configurable
  });

  useEffect(() => {
    if (error) {
      toast.error("Failed to load events");
    }
  }, [error]);

  if (isLoading) {
    return (
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {Array.from({ length: 6 }).map((_, i) => (
          <Card key={i} className="animate-pulse">
            <CardHeader>
              <div className="h-4 bg-gray-200 rounded w-3/4 mb-2"></div>
              <div className="h-3 bg-gray-200 rounded w-1/2"></div>
            </CardHeader>
            <CardContent>
              <div className="h-20 bg-gray-200 rounded"></div>
            </CardContent>
          </Card>
        ))}
      </div>
    );
  }

  if (!eventsResponse?.success || !eventsResponse.data) {
    return (
      <div className="text-center py-8">
        <p className="text-muted-foreground">No events available or failed to load events.</p>
      </div>
    );
  }

  const events = eventsResponse.data;

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

  const getCategoryFromName = (name: string, description?: string) => {
    const searchText = `${name} ${description || ''}`.toLowerCase();
    
    // Check each category for keyword matches
    for (const [category, keywords] of Object.entries(EVENT_CATEGORIES)) {
      if (keywords.some(keyword => searchText.includes(keyword))) {
        return category;
      }
    }
    
    return 'Event'; // Default category
  };

  const getEventEmoji = (category: string) => {
    return EVENT_EMOJIS[category as keyof typeof EVENT_EMOJIS] || EVENT_EMOJIS['Event'];
  };

  // Calculate a simple rating based on available tickets (simulated)
  const calculateEventRating = (totalTickets: number, availableTickets: number) => {
    const soldPercentage = ((totalTickets - availableTickets) / totalTickets) * 100;
    // Higher sold percentage = higher rating (popular events)
    const baseRating = 3.5 + (soldPercentage / 100) * 1.5;
    return Math.min(5.0, Math.max(3.0, baseRating)).toFixed(1);
  };

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      {events.map((event) => {
        const category = getCategoryFromName(event.name, event.description);
        const emoji = getEventEmoji(category);
        const rating = calculateEventRating(event.total_tickets, event.available_tickets || event.total_tickets);
        
        return (
          <Card 
            key={event.id} 
            className="group hover:shadow-xl transition-all duration-300 hover:-translate-y-1 cursor-pointer border-0 shadow-md bg-white/70 backdrop-blur-sm"
            onClick={() => onEventSelect(event)}
          >
            <CardHeader className="pb-3">
              <div className="flex items-start justify-between">
                <div className="text-4xl mb-2">{emoji}</div>
                <Badge 
                  variant="secondary" 
                  className="bg-gradient-to-r from-blue-500 to-purple-600 text-white border-0"
                >
                  {category}
                </Badge>
              </div>
              <CardTitle className="text-xl group-hover:text-primary transition-colors">
                {event.name}
              </CardTitle>
              <div className="space-y-2 mt-2">
                <div className="flex items-center space-x-4 text-sm">
                  <span className="flex items-center space-x-1">
                    <Calendar className="h-4 w-4 text-blue-600" />
                    <span>{formatDate(event.start_time)}</span>
                  </span>
                  <span className="flex items-center space-x-1">
                    <Clock className="h-4 w-4 text-blue-600" />
                    <span>{formatTime(event.start_time)}</span>
                  </span>
                </div>
                <div className="flex items-center space-x-1 text-sm">
                  <MapPin className="h-4 w-4 text-blue-600" />
                  <span>{event.venue}</span>
                </div>
              </div>
            </CardHeader>
            
            <CardContent>
              <div className="flex items-center justify-between mb-4">
                <div className="flex items-center space-x-3">
                  <div className="flex items-center space-x-1">
                    <Star className="h-4 w-4 fill-yellow-400 text-yellow-400" />
                    <span className="text-sm font-medium">{rating}</span>
                  </div>
                  <div className="flex items-center space-x-1 text-sm text-muted-foreground">
                    <Users className="h-4 w-4" />
                    <span>{event.total_tickets}</span>
                  </div>
                </div>
                <div className="text-right">
                  <div className="text-2xl font-bold text-primary">${event.price}</div>
                  <div className="text-xs text-muted-foreground">per ticket</div>
                </div>
              </div>
              
              <div className="flex items-center justify-between">
                <div className="text-sm">
                  <span className="text-green-600 font-medium">{event.available_tickets || event.total_tickets}</span>
                  <span className="text-muted-foreground"> seats left</span>
                </div>
                <Button 
                  size="sm" 
                  className="bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 transition-all duration-300"
                >
                  Select Seats
                </Button>
              </div>
            </CardContent>
          </Card>
        );
      })}
    </div>
  );
};

export default EventGrid;
