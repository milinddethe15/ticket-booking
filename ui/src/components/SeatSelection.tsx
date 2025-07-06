import { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Monitor, Zap, Users, CheckCircle, RefreshCw } from "lucide-react";
import { toast } from "sonner";
import { useQuery } from "@tanstack/react-query";
import { apiService, Ticket } from "@/services/api";
import { getConfig, getSectionForRow, calculateSeatPrice, generateSeatNumber } from "@/lib/config";

interface SeatSelectionProps {
  eventId: string;
  onSeatSelect: (seats: string[]) => void;
  selectedSeats: string[];
  eventPrice: number;
}

type SeatStatus = 'available' | 'selected' | 'occupied' | 'loading';

interface Seat {
  id: string;
  row: string;
  number: number;
  status: SeatStatus;
  price: number;
  section: string;
  ticketId?: number;
  seatNo?: string;
}

const SeatSelection = ({ eventId, onSeatSelect, selectedSeats, eventPrice }: SeatSelectionProps) => {
  const [seats, setSeats] = useState<Seat[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [lastUpdate, setLastUpdate] = useState<Date>(new Date());
  const config = getConfig();

  // Fetch available tickets from the backend with polling
  const { data: ticketsResponse, isLoading: isLoadingTickets, refetch: refetchAvailable } = useQuery({
    queryKey: ['tickets', eventId],
    queryFn: () => apiService.getAvailableTickets(parseInt(eventId), config.venue.totalSeats),
    enabled: !!eventId,
    refetchInterval: config.refreshIntervalSeconds * 1000,
    refetchIntervalInBackground: true,
  });

  // Also fetch all tickets to know which ones are sold/reserved
  const { data: allTicketsResponse, refetch: refetchAll } = useQuery({
    queryKey: ['all-tickets', eventId],
    queryFn: () => apiService.getAllTickets(parseInt(eventId), config.venue.totalSeats),
    enabled: !!eventId,
    refetchInterval: config.refreshIntervalSeconds * 1000,
    refetchIntervalInBackground: true,
  });

  // Generate seat layout based on actual ticket data
  useEffect(() => {
    if (!allTicketsResponse?.success || !allTicketsResponse.data) {
      return;
    }

    const allTickets = allTicketsResponse.data as Ticket[];
    const generateSeats = () => {
      const seatData: Seat[] = [];
      const { rows, seatsPerRow } = config.venue;
      
      // Create a map of all tickets by their seat number
      const ticketsMap = new Map<string, Ticket>();
      allTickets.forEach(ticket => {
        ticketsMap.set(ticket.seat_no, ticket);
      });

      console.log(`Event ${eventId}: Found ${allTickets.length} tickets, seat range: ${allTickets[0]?.seat_no} to ${allTickets[allTickets.length - 1]?.seat_no}`);

      rows.forEach((row, rowIndex) => {
        for (let seatNum = 1; seatNum <= seatsPerRow; seatNum++) {
          const section = getSectionForRow(rowIndex);
          const seatPrice = calculateSeatPrice(eventPrice, rowIndex);
          const visualSeatId = `${row}${seatNum}`;
          
          // Generate backend seat number using our mapping function
          const backendSeatNo = generateSeatNumber(rowIndex, seatNum - 1);
          
          // Check ticket status in the backend
          const ticket = ticketsMap.get(backendSeatNo);
          let status: SeatStatus = 'occupied'; // Default to occupied if ticket doesn't exist
          
          if (ticket) {
            switch (ticket.status) {
              case 'available':
                status = 'available';
                break;
              case 'locked':
              case 'reserved':
              case 'sold':
                status = 'occupied';
                break;
              default:
                status = 'occupied';
            }
          }
          
          seatData.push({
            id: visualSeatId,
            row,
            number: seatNum,
            status,
            price: seatPrice,
            section: section.name,
            ticketId: ticket?.id,
            seatNo: backendSeatNo,
          });
        }
      });
      
      setSeats(seatData);
      setLastUpdate(new Date());
    };

    generateSeats();
  }, [eventId, eventPrice, allTicketsResponse, config]);

  // Handle manual refresh
  const handleRefresh = async () => {
    await Promise.all([refetchAvailable(), refetchAll()]);
    toast.success("Seat availability refreshed");
  };

  const handleSeatClick = async (seatId: string) => {
    const seat = seats.find(s => s.id === seatId);
    if (!seat || seat.status === 'occupied') {
      toast.error("This seat is no longer available");
      return;
    }

    // Check if seat is still available by refetching
    setIsLoading(true);
    setSeats(prev => prev.map(s => 
      s.id === seatId ? { ...s, status: 'loading' as SeatStatus } : s
    ));

    try {
      // Refresh both available tickets and all tickets data
      await Promise.all([refetchAvailable(), refetchAll()]);
      
      // Re-check availability after refresh
      const updatedTickets = (await apiService.getAvailableTickets(parseInt(eventId), config.venue.totalSeats)).data as Ticket[];
      const isStillAvailable = updatedTickets.some(ticket => ticket.seat_no === seat.seatNo);
      
      if (!isStillAvailable) {
        toast.error("Seat was just booked by another user. Please select a different seat.");
        setSeats(prev => prev.map(s => 
          s.id === seatId ? { ...s, status: 'occupied' as SeatStatus } : s
        ));
        setIsLoading(false);
        return;
      }

      const isCurrentlySelected = selectedSeats.includes(seatId);
      
      if (isCurrentlySelected) {
        // Deselecting seat - unlock it
        try {
          await apiService.unlockSeat(parseInt(eventId), seat.seatNo!);
          const newSelectedSeats = selectedSeats.filter(id => id !== seatId);
          setSeats(prev => prev.map(s => 
            s.id === seatId ? { ...s, status: 'available' as SeatStatus } : s
          ));
          onSeatSelect(newSelectedSeats);
          toast.success("Seat deselected and unlocked");
        } catch (error) {
          toast.error("Failed to unlock seat");
        }
      } else {
        // Check if we're at the maximum seats limit
        if (selectedSeats.length >= config.maxSeatsPerBooking) {
          toast.error(`Maximum ${config.maxSeatsPerBooking} seats allowed per booking`);
          setSeats(prev => prev.map(s => 
            s.id === seatId ? { ...s, status: 'available' as SeatStatus } : s
          ));
          setIsLoading(false);
          return;
        }

        // Selecting seat - lock it
        try {
          console.log(`Attempting to lock seat: ${seat.seatNo} for event ${eventId}`);
          const lockResponse = await apiService.lockSeat(parseInt(eventId), seat.seatNo!);
          console.log(`Lock response:`, lockResponse);
          if (lockResponse.success) {
            const newSelectedSeats = [...selectedSeats, seatId];
            setSeats(prev => prev.map(s => 
              s.id === seatId ? { ...s, status: 'selected' as SeatStatus } : s
            ));
            onSeatSelect(newSelectedSeats);
            toast.success(`Seat locked! Complete booking within ${config.seatLockDurationMinutes} minutes.`);
          } else {
            throw new Error(lockResponse.error || "Failed to lock seat");
          }
        } catch (error) {
          console.error("Seat lock error:", error);
          const errorMessage = (error as Error)?.message || "Seat was just taken by another user";
          toast.error(`Cannot select seat: ${errorMessage}`);
          // Refresh to get latest state
          await Promise.all([refetchAvailable(), refetchAll()]);
        }
      }
    } catch (error) {
      toast.error("Failed to verify seat availability. Please try again.");
      setSeats(prev => prev.map(s => 
        s.id === seatId ? { ...s, status: 'available' as SeatStatus } : s
      ));
    } finally {
      setIsLoading(false);
    }
  };

  const getSeatColor = (status: SeatStatus) => {
    switch (status) {
      case 'available': return 'bg-green-100 hover:bg-green-200 border-green-300 text-green-800';
      case 'selected': return 'bg-blue-500 text-white border-blue-600 seat-pulse';
      case 'occupied': return 'bg-red-100 border-red-300 text-red-800 cursor-not-allowed opacity-60';
      case 'loading': return 'bg-yellow-100 border-yellow-300 text-yellow-800 animate-pulse';
    }
  };

  const totalPrice = selectedSeats.reduce((sum, seatId) => {
    const seat = seats.find(s => s.id === seatId);
    return sum + (seat?.price || 0);
  }, 0);

  const sectionCounts = seats.reduce((acc, seat) => {
    if (seat.status === 'available') {
      acc[seat.section] = (acc[seat.section] || 0) + 1;
    }
    return acc;
  }, {} as Record<string, number>);

  if (isLoadingTickets) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <RefreshCw className="h-8 w-8 animate-spin mx-auto mb-2" />
          <p>Loading seat availability...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
      {/* Seat Map */}
      <div className="lg:col-span-3">
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle className="flex items-center space-x-2">
                <Monitor className="h-5 w-5" />
                <span>Select Your Seats</span>
              </CardTitle>
              <div className="flex items-center space-x-4">
                <div className="flex items-center space-x-2 text-sm text-muted-foreground">
                  <Zap className="h-4 w-4 text-yellow-500" />
                  <span>Real-time availability</span>
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleRefresh}
                  disabled={isLoading}
                >
                  <RefreshCw className={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
                </Button>
              </div>
            </div>
            <p className="text-sm text-muted-foreground">
              Last updated: {lastUpdate.toLocaleTimeString()}
            </p>
          </CardHeader>
          
          <CardContent>
            {/* Stage */}
            <div className="text-center mb-8">
              <div className="booking-gradient text-white py-3 px-6 rounded-lg inline-block text-sm font-medium">
                🎪 STAGE
              </div>
            </div>

            {/* Seat Grid */}
            <div className="space-y-2">
              {config.venue.rows.map((row) => (
                <div key={row} className="flex items-center justify-center space-x-1">
                  <div className="w-6 text-center font-medium text-sm text-muted-foreground">
                    {row}
                  </div>
                  <div className="flex space-x-1">
                    {Array.from({ length: config.venue.seatsPerRow }, (_, i) => {
                      const seatId = `${row}${i + 1}`;
                      const seat = seats.find(s => s.id === seatId);
                      if (!seat) return null;

                      return (
                        <button
                          key={seatId}
                          onClick={() => handleSeatClick(seatId)}
                          disabled={seat.status === 'occupied' || seat.status === 'loading' || isLoading}
                          className={`w-8 h-8 rounded text-xs font-medium transition-all duration-200 border-2 ${getSeatColor(seat.status)}`}
                          title={`${seat.section} - $${seat.price}${seat.seatNo ? ` (${seat.seatNo})` : ''}`}
                        >
                          {seat.status === 'loading' ? '⏳' : seat.number}
                        </button>
                      );
                    })}
                  </div>
                </div>
              ))}
            </div>

            {/* Legend */}
            <div className="flex justify-center items-center space-x-6 mt-8 text-sm">
              <div className="flex items-center space-x-2">
                <div className="w-4 h-4 bg-green-100 border-2 border-green-300 rounded"></div>
                <span>Available</span>
              </div>
              <div className="flex items-center space-x-2">
                <div className="w-4 h-4 bg-blue-500 border-2 border-blue-600 rounded"></div>
                <span>Selected</span>
              </div>
              <div className="flex items-center space-x-2">
                <div className="w-4 h-4 bg-red-100 border-2 border-red-300 rounded opacity-60"></div>
                <span>Occupied</span>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Booking Summary */}
      <div className="space-y-4">
        {/* Section Availability */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Section Availability</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {Object.entries(sectionCounts).map(([section, count]) => (
              <div key={section} className="flex items-center justify-between">
                <span className="font-medium">{section}</span>
                <Badge variant="outline" className="text-green-600">
                  {count} left
                </Badge>
              </div>
            ))}
          </CardContent>
        </Card>

        {/* Selection Summary */}
        {selectedSeats.length > 0 && (
          <Card>
            <CardHeader>
              <CardTitle className="text-lg flex items-center space-x-2">
                <CheckCircle className="h-5 w-5 text-green-600" />
                <span>Your Selection</span>
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="space-y-2">
                {selectedSeats.map(seatId => {
                  const seat = seats.find(s => s.id === seatId);
                  if (!seat) return null;
                  
                  return (
                    <div key={seatId} className="flex items-center justify-between text-sm">
                      <span>{seat.section} - Row {seat.row}, Seat {seat.number}</span>
                      <span className="font-medium">${seat.price}</span>
                    </div>
                  );
                })}
              </div>
              
              <div className="border-t pt-3">
                <div className="flex items-center justify-between font-bold text-lg">
                  <span>Total ({selectedSeats.length} seats)</span>
                  <span className="text-primary">${totalPrice}</span>
                </div>
              </div>

              <Button 
                className="w-full bg-gradient-to-r from-green-600 to-blue-600 hover:from-green-700 hover:to-blue-700"
                size="lg"
              >
                Continue to Checkout
              </Button>
            </CardContent>
          </Card>
        )}

        {/* Live Stats */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg flex items-center space-x-2">
              <Users className="h-5 w-5 text-blue-600" />
              <span>Live Stats</span>
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-2 text-sm">
            <div className="flex justify-between">
              <span>Available tickets:</span>
              <span className="font-medium text-green-600">
                {ticketsResponse?.data?.length || 0}
              </span>
            </div>
            <div className="flex justify-between">
              <span>Last updated:</span>
              <span className="font-medium">{lastUpdate.toLocaleTimeString()}</span>
            </div>
            <div className="flex justify-between">
              <span>Auto-refresh:</span>
              <span className="font-medium text-blue-600">Every {config.refreshIntervalSeconds}s</span>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
};

export default SeatSelection;
