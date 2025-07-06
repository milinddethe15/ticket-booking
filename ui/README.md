# Ticket Booking UI

A modern, responsive React application for high-performance ticket booking with real-time seat selection and dynamic configuration.

## ✨ Features

- **Fully Dynamic Configuration**: All UI behavior is configurable via environment variables
- **Real-time Seat Selection**: Live seat availability updates with WebSocket-like polling
- **Responsive Design**: Works seamlessly on desktop, tablet, and mobile devices
- **Modern UI Components**: Built with shadcn/ui and Tailwind CSS
- **Optimistic Updates**: Fast UI feedback with backend validation
- **Error Handling**: Graceful error handling with user-friendly messages
- **Accessibility**: Full keyboard navigation and screen reader support

## 🔧 Dynamic Configuration

The application is now completely configurable without code changes. All hardcoded values have been removed and replaced with environment variables.

### Configurable Elements:
- **API endpoints** and base URLs
- **Venue layout** (rows, seats per row, total seats)
- **Seat sections** and pricing multipliers
- **Booking parameters** (timeouts, limits, fees)
- **UI behavior** (refresh rates, animations)
- **User settings** (default user ID, preferences)

See [CONFIGURATION.md](./CONFIGURATION.md) for detailed configuration options.

## 🚀 Getting Started

### Prerequisites
- Node.js 18+ and npm/yarn
- Running backend server (see `/server` directory)

### Installation
```bash
npm install
```

### Configuration
1. Copy the example environment file:
```bash
cp .env.example .env
```

2. Configure your environment variables in `.env`:
```bash
# API Configuration
VITE_API_BASE_URL=http://localhost:8080/api/v1

# Venue Configuration (customize for your venue)
VITE_VENUE_ROWS=A,B,C,D,E,F,G,H
VITE_VENUE_SEATS_PER_ROW=12
VITE_VENUE_TOTAL_SEATS=96

# Booking Configuration
VITE_SERVICE_FEE_PERCENTAGE=0.08
VITE_MAX_SEATS_PER_BOOKING=8
# ... see CONFIGURATION.md for full options
```

### Development
```bash
npm run dev
```

### Production Build
```bash
npm run build
npm run preview
```

## 📁 Project Structure

```
src/
├── components/
│   ├── ui/              # Reusable UI components (shadcn/ui)
│   ├── EventGrid.tsx    # Dynamic event listing with smart categorization
│   ├── SeatSelection.tsx # Configurable seat map with real-time updates
│   └── BookingFlow.tsx  # Multi-step booking process
├── lib/
│   ├── config.ts        # Centralized configuration system
│   └── utils.ts         # Utility functions
├── pages/
│   ├── Index.tsx        # Main application page
│   └── NotFound.tsx     # 404 page
├── services/
│   └── api.ts          # API service layer
└── hooks/              # Custom React hooks
```

## 🎯 Key Components

### EventGrid
- **Smart categorization** based on event names and descriptions
- **Dynamic ratings** calculated from ticket sales
- **Configurable display** options
- **Real-time availability** updates

### SeatSelection
- **Configurable venue layouts** (any number of rows/seats)
- **Dynamic section pricing** (VIP, Premium, Standard)
- **Real-time seat locking** with visual feedback
- **Optimistic UI updates** with backend validation

### BookingFlow
- **Multi-step booking process** with validation
- **Dynamic fee calculation** based on configuration
- **Real-time countdown timers**
- **Secure payment simulation**

## 🛠️ Technology Stack

- **React 18** with TypeScript
- **Vite** for fast development and building
- **Tailwind CSS** for styling
- **shadcn/ui** for components
- **React Query** for data fetching
- **React Router** for navigation
- **Sonner** for notifications

## 🎨 Customization

### Venue Layouts
The application supports any venue layout through configuration:

```bash
# Small Theater (4 rows, 10 seats each)
VITE_VENUE_ROWS=A,B,C,D
VITE_VENUE_SEATS_PER_ROW=10
VITE_VENUE_TOTAL_SEATS=40

# Large Stadium (20 rows, 50 seats each)
VITE_VENUE_ROWS=A,B,C,D,E,F,G,H,I,J,K,L,M,N,O,P,Q,R,S,T
VITE_VENUE_SEATS_PER_ROW=50
VITE_VENUE_TOTAL_SEATS=1000
```

### Pricing Sections
Configure different pricing tiers:

```bash
VITE_VIP_PRICE_MULTIPLIER=2.0      # VIP seats cost 2x base price
VITE_PREMIUM_PRICE_MULTIPLIER=1.5   # Premium seats cost 1.5x base price
VITE_STANDARD_PRICE_MULTIPLIER=1.0  # Standard seats cost base price
```

### Booking Behavior
Customize booking timeouts and limits:

```bash
VITE_SEAT_LOCK_DURATION_MINUTES=5    # Hold seats for 5 minutes
VITE_BOOKING_EXPIRATION_MINUTES=20   # Complete booking within 20 minutes
VITE_MAX_SEATS_PER_BOOKING=12        # Allow up to 12 seats per booking
```

## 🔐 Security

- Environment variables starting with `VITE_` are publicly accessible
- Never put sensitive data in client-side configuration
- Use proper authentication in production (replace `VITE_DEFAULT_USER_ID`)
- Implement proper session management