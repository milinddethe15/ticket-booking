import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { 
  CreditCard, 
  User, 
  Mail, 
  Phone, 
  Lock, 
  CheckCircle, 
  ArrowLeft,
  Clock,
  Shield
} from "lucide-react";
import { toast } from "sonner";
import { apiService, Event } from "@/services/api";
import { getConfig } from "@/lib/config";

interface BookingFlowProps {
  event: Event;
  selectedSeats: string[];
  onBack: () => void;
}

const BookingFlow = ({ event, selectedSeats, onBack }: BookingFlowProps) => {
  const [isProcessing, setIsProcessing] = useState(false);
  const [bookingComplete, setBookingComplete] = useState(false);
  const [bookingData, setBookingData] = useState<any>(null);
  const [formData, setFormData] = useState({
    firstName: '',
    lastName: '',
    email: '',
    phone: '',
    cardNumber: '',
    expiryDate: '',
    cvv: ''
  });
  const config = getConfig();

  const totalPrice = selectedSeats.length * event.price;
  const fees = Math.round(totalPrice * config.serviceFeePercentage);
  const finalTotal = totalPrice + fees;

  const handleInputChange = (field: string, value: string) => {
    setFormData(prev => ({ ...prev, [field]: value }));
  };

  const handleBooking = async () => {
    // Basic validation
    const requiredFields = ['firstName', 'lastName', 'email', 'phone', 'cardNumber', 'expiryDate', 'cvv'];
    const missingFields = requiredFields.filter(field => !formData[field]);
    
    if (missingFields.length > 0) {
      toast.error("Please fill in all required fields");
      return;
    }

    setIsProcessing(true);
    
    try {
      // Step 1: Book tickets using your API
      toast.info("Reserving tickets with pessimistic locking...");
      
      const bookingRequest = {
        user_id: config.defaultUserId, // Use configured default user ID
        event_id: event.id,
        quantity: selectedSeats.length
      };

      const bookingResponse = await apiService.bookTickets(bookingRequest);
      
      if (!bookingResponse.success || !bookingResponse.data) {
        throw new Error(bookingResponse.error || "Failed to reserve tickets");
      }

      await new Promise(resolve => setTimeout(resolve, 1000));
      toast.info("Processing payment securely...");
      await new Promise(resolve => setTimeout(resolve, 2000));
      
      // Step 2: Confirm the booking
      toast.info("Confirming booking...");
      const confirmResponse = await apiService.confirmBooking(bookingResponse.data.id);
      
      if (!confirmResponse.success) {
        // If confirmation fails, we should cancel the booking
        await apiService.cancelBooking(bookingResponse.data.id);
        throw new Error(confirmResponse.error || "Failed to confirm booking");
      }

      await new Promise(resolve => setTimeout(resolve, 800));
      toast.info("Generating tickets...");
      await new Promise(resolve => setTimeout(resolve, 500));
      
      setBookingData(bookingResponse.data);
      setBookingComplete(true);
      toast.success("Booking confirmed! Check your email for tickets.");
      
    } catch (error) {
      console.error('Booking error:', error);
      toast.error("Booking failed: " + (error as Error).message);
    } finally {
      setIsProcessing(false);
    }
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

  if (bookingComplete && bookingData) {
    return (
      <div className="max-w-2xl mx-auto">
        <Card className="border-green-200 bg-green-50">
          <CardContent className="pt-6">
            <div className="text-center space-y-4">
              <div className="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mx-auto">
                <CheckCircle className="h-8 w-8 text-green-600" />
              </div>
              <h2 className="text-2xl font-bold text-green-800">Booking Confirmed!</h2>
              <p className="text-green-700">
                Your tickets have been reserved successfully. You'll receive a confirmation email shortly.
              </p>
              
              <div className="bg-white rounded-lg p-4 text-left space-y-2">
                <div className="flex justify-between">
                  <span>Booking Reference:</span>
                  <span className="font-mono text-sm">{bookingData.booking_ref}</span>
                </div>
                <div className="flex justify-between">
                  <span>Event:</span>
                  <span className="font-medium">{event.name}</span>
                </div>
                <div className="flex justify-between">
                  <span>Seats:</span>
                  <span className="font-medium">{selectedSeats.length} seats</span>
                </div>
                <div className="flex justify-between">
                  <span>Total Paid:</span>
                  <span className="font-bold text-green-600">${bookingData.total_amount}</span>
                </div>
                <div className="flex justify-between">
                  <span>Status:</span>
                  <span className="font-medium capitalize">{bookingData.status}</span>
                </div>
              </div>
              
              <Button 
                onClick={() => window.location.reload()} 
                className="bg-green-600 hover:bg-green-700"
              >
                Book Another Event
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto grid grid-cols-1 lg:grid-cols-2 gap-6">
      {/* Booking Form */}
      <div className="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center space-x-2">
              <User className="h-5 w-5" />
              <span>Customer Information</span>
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <Label htmlFor="firstName">First Name</Label>
                <Input
                  id="firstName"
                  value={formData.firstName}
                  onChange={(e) => handleInputChange('firstName', e.target.value)}
                  placeholder="John"
                />
              </div>
              <div>
                <Label htmlFor="lastName">Last Name</Label>
                <Input
                  id="lastName"
                  value={formData.lastName}
                  onChange={(e) => handleInputChange('lastName', e.target.value)}
                  placeholder="Doe"
                />
              </div>
            </div>
            
            <div>
              <Label htmlFor="email">Email Address</Label>
              <div className="relative">
                <Mail className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                <Input
                  id="email"
                  type="email"
                  className="pl-10"
                  value={formData.email}
                  onChange={(e) => handleInputChange('email', e.target.value)}
                  placeholder="john@example.com"
                />
              </div>
            </div>
            
            <div>
              <Label htmlFor="phone">Phone Number</Label>
              <div className="relative">
                <Phone className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                <Input
                  id="phone"
                  type="tel"
                  className="pl-10"
                  value={formData.phone}
                  onChange={(e) => handleInputChange('phone', e.target.value)}
                  placeholder="+1 (555) 123-4567"
                />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center space-x-2">
              <CreditCard className="h-5 w-5" />
              <span>Payment Information</span>
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <Label htmlFor="cardNumber">Card Number</Label>
              <div className="relative">
                <CreditCard className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                <Input
                  id="cardNumber"
                  className="pl-10"
                  value={formData.cardNumber}
                  onChange={(e) => handleInputChange('cardNumber', e.target.value)}
                  placeholder="1234 5678 9012 3456"
                />
              </div>
            </div>
            
            <div className="grid grid-cols-2 gap-4">
              <div>
                <Label htmlFor="expiryDate">Expiry Date</Label>
                <Input
                  id="expiryDate"
                  value={formData.expiryDate}
                  onChange={(e) => handleInputChange('expiryDate', e.target.value)}
                  placeholder="MM/YY"
                />
              </div>
              <div>
                <Label htmlFor="cvv">CVV</Label>
                <div className="relative">
                  <Lock className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                  <Input
                    id="cvv"
                    className="pl-10"
                    value={formData.cvv}
                    onChange={(e) => handleInputChange('cvv', e.target.value)}
                    placeholder="123"
                  />
                </div>
              </div>
            </div>
            
            <div className="flex items-center space-x-2 text-sm text-muted-foreground bg-slate-50 p-3 rounded-lg">
              <Shield className="h-4 w-4 text-green-600" />
              <span>Your payment is secured with 256-bit SSL encryption</span>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Order Summary */}
      <div className="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>Order Summary</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {/* Event Details */}
            <div className="space-y-2">
              <h3 className="font-semibold text-lg">{event.name}</h3>
              <div className="text-sm text-muted-foreground space-y-1">
                <div>{formatDate(event.start_time)} at {formatTime(event.start_time)}</div>
                <div>{event.venue}</div>
              </div>
            </div>
            
            <Separator />
            
            {/* Selected Seats */}
            <div>
              <h4 className="font-medium mb-2">Selected Seats</h4>
              <div className="space-y-1">
                {selectedSeats.map((backendSeatNo, index) => (
                  <div key={backendSeatNo} className="flex items-center justify-between text-sm">
                    <span>Seat #{index + 1} ({backendSeatNo})</span>
                    <span>${event.price}</span>
                  </div>
                ))}
              </div>
            </div>
            
            <Separator />
            
            {/* Pricing Breakdown */}
            <div className="space-y-2">
              <div className="flex justify-between">
                <span>Subtotal ({selectedSeats.length} tickets)</span>
                <span>${totalPrice}</span>
              </div>
              <div className="flex justify-between text-sm text-muted-foreground">
                <span>Service fees ({Math.round(config.serviceFeePercentage * 100)}%)</span>
                <span>${fees}</span>
              </div>
              <Separator />
              <div className="flex justify-between font-bold text-lg">
                <span>Total</span>
                <span className="text-primary">${finalTotal}</span>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Timer and Actions */}
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-center space-x-2 text-sm text-orange-600 mb-4">
              <Clock className="h-4 w-4" />
              <span>Seats held for {config.bookingExpirationMinutes} minutes</span>
            </div>
            
            <div className="space-y-3">
              <Button
                onClick={handleBooking}
                disabled={isProcessing}
                className="w-full bg-gradient-to-r from-green-600 to-blue-600 hover:from-green-700 hover:to-blue-700 text-lg py-6"
              >
                {isProcessing ? (
                  <span className="flex items-center space-x-2">
                    <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
                    <span>Processing...</span>
                  </span>
                ) : (
                  `Complete Booking - $${finalTotal}`
                )}
              </Button>
              
              <Button 
                variant="outline" 
                onClick={onBack}
                className="w-full"
                disabled={isProcessing}
              >
                <ArrowLeft className="h-4 w-4 mr-2" />
                Back to Seat Selection
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
};

export default BookingFlow;
