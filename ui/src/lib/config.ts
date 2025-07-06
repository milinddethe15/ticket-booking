// Configuration for the ticket booking application
export interface SeatSection {
  name: string;
  priceMultiplier: number;
  color: string;
  rows: number[];
}

export interface VenueConfig {
  rows: string[];
  seatsPerRow: number;
  sections: SeatSection[];
  totalSeats: number;
}

export interface AppConfig {
  // API Configuration
  apiBaseUrl: string;
  
  // Booking Configuration
  defaultUserId: number; // In production, this would come from authentication
  serviceFeePercentage: number;
  seatLockDurationMinutes: number;
  bookingExpirationMinutes: number;
  
  // UI Configuration
  refreshIntervalSeconds: number;
  maxSeatsPerBooking: number;
  
  // Venue Configuration
  venue: VenueConfig;
}

// Default configuration - can be overridden by environment variables
export const defaultConfig: AppConfig = {
  apiBaseUrl: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1',
  
  // Default user ID - in production this would come from authentication
  defaultUserId: parseInt(import.meta.env.VITE_DEFAULT_USER_ID || '1'),
  
  // Service fee percentage (8% by default)
  serviceFeePercentage: parseFloat(import.meta.env.VITE_SERVICE_FEE_PERCENTAGE || '0.08'),
  
  // Seat lock duration in minutes
  seatLockDurationMinutes: parseInt(import.meta.env.VITE_SEAT_LOCK_DURATION_MINUTES || '3'),
  
  // Booking expiration in minutes
  bookingExpirationMinutes: parseInt(import.meta.env.VITE_BOOKING_EXPIRATION_MINUTES || '15'),
  
  // UI refresh interval in seconds
  refreshIntervalSeconds: parseInt(import.meta.env.VITE_REFRESH_INTERVAL_SECONDS || '3'),
  
  // Maximum seats per booking
  maxSeatsPerBooking: parseInt(import.meta.env.VITE_MAX_SEATS_PER_BOOKING || '8'),
  
  // Venue configuration
  venue: {
    rows: (import.meta.env.VITE_VENUE_ROWS || 'A,B,C,D,E,F,G,H').split(','),
    seatsPerRow: parseInt(import.meta.env.VITE_VENUE_SEATS_PER_ROW || '12'),
    totalSeats: parseInt(import.meta.env.VITE_VENUE_TOTAL_SEATS || '96'),
    sections: [
      {
        name: 'VIP',
        priceMultiplier: parseFloat(import.meta.env.VITE_VIP_PRICE_MULTIPLIER || '1.5'),
        color: 'from-amber-500 to-orange-600',
        rows: [0, 1] // First 2 rows
      },
      {
        name: 'Premium',
        priceMultiplier: parseFloat(import.meta.env.VITE_PREMIUM_PRICE_MULTIPLIER || '1.2'),
        color: 'from-blue-500 to-purple-600',
        rows: [2, 3, 4] // Next 3 rows
      },
      {
        name: 'Standard',
        priceMultiplier: parseFloat(import.meta.env.VITE_STANDARD_PRICE_MULTIPLIER || '1.0'),
        color: 'from-green-500 to-blue-500',
        rows: [5, 6, 7] // Last 3 rows
      }
    ]
  }
};

// Get configuration with environment variable overrides
export const getConfig = (): AppConfig => {
  return defaultConfig;
};

// Helper functions
export const getSectionForRow = (rowIndex: number): SeatSection => {
  const config = getConfig();
  for (const section of config.venue.sections) {
    if (section.rows.includes(rowIndex)) {
      return section;
    }
  }
  // Default to standard if no section found
  return config.venue.sections[config.venue.sections.length - 1];
};

export const calculateSeatPrice = (basePrice: number, rowIndex: number): number => {
  const section = getSectionForRow(rowIndex);
  return Math.round(basePrice * section.priceMultiplier);
};

export const generateSeatNumber = (rowIndex: number, seatIndex: number): string => {
  const config = getConfig();
  const totalSeatIndex = rowIndex * config.venue.seatsPerRow + seatIndex + 1;
  return `S${totalSeatIndex.toString().padStart(3, '0')}`;
}; 