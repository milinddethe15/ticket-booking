# Configuration Options

The ticket booking application supports dynamic configuration through environment variables. This allows you to customize the application behavior without modifying the code.

## Environment Variables

Create a `.env` file in the `ui` directory with the following variables:

### API Configuration
- `VITE_API_BASE_URL` - Base URL for the API server (default: `http://localhost:8080/api/v1`)

### Booking Configuration
- `VITE_DEFAULT_USER_ID` - Default user ID for bookings (default: `1`)
  - In production, this should be replaced with proper authentication
- `VITE_SERVICE_FEE_PERCENTAGE` - Service fee percentage (default: `0.08` = 8%)
- `VITE_SEAT_LOCK_DURATION_MINUTES` - How long seats are locked during selection (default: `3`)
- `VITE_BOOKING_EXPIRATION_MINUTES` - How long users have to complete payment (default: `15`)

### UI Configuration
- `VITE_REFRESH_INTERVAL_SECONDS` - How often to refresh seat availability (default: `3`)
- `VITE_MAX_SEATS_PER_BOOKING` - Maximum number of seats per booking (default: `8`)

### Venue Configuration
- `VITE_VENUE_ROWS` - Comma-separated list of row names (default: `A,B,C,D,E,F,G,H`)
- `VITE_VENUE_SEATS_PER_ROW` - Number of seats per row (default: `12`)
- `VITE_VENUE_TOTAL_SEATS` - Total number of seats in the venue (default: `96`)

### Pricing Configuration
- `VITE_VIP_PRICE_MULTIPLIER` - Price multiplier for VIP section (default: `1.5`)
- `VITE_PREMIUM_PRICE_MULTIPLIER` - Price multiplier for Premium section (default: `1.2`)
- `VITE_STANDARD_PRICE_MULTIPLIER` - Price multiplier for Standard section (default: `1.0`)

## Example .env File

```bash
# API Configuration
VITE_API_BASE_URL=http://localhost:8080/api/v1

# Booking Configuration
VITE_DEFAULT_USER_ID=1
VITE_SERVICE_FEE_PERCENTAGE=0.08
VITE_SEAT_LOCK_DURATION_MINUTES=3
VITE_BOOKING_EXPIRATION_MINUTES=15

# UI Configuration
VITE_REFRESH_INTERVAL_SECONDS=3
VITE_MAX_SEATS_PER_BOOKING=8

# Venue Configuration
VITE_VENUE_ROWS=A,B,C,D,E,F,G,H
VITE_VENUE_SEATS_PER_ROW=12
VITE_VENUE_TOTAL_SEATS=96

# Pricing Configuration
VITE_VIP_PRICE_MULTIPLIER=1.5
VITE_PREMIUM_PRICE_MULTIPLIER=1.2
VITE_STANDARD_PRICE_MULTIPLIER=1.0
```

## Seat Section Configuration

The application automatically assigns seats to sections based on row positions:
- **VIP Section**: First 2 rows (configurable via code)
- **Premium Section**: Next 3 rows (configurable via code)
- **Standard Section**: Remaining rows (configurable via code)

To modify section assignments, update the `venue.sections` configuration in `src/lib/config.ts`.

## Venue Layout Examples

### Small Venue (48 seats)
```bash
VITE_VENUE_ROWS=A,B,C,D
VITE_VENUE_SEATS_PER_ROW=12
VITE_VENUE_TOTAL_SEATS=48
```

### Large Venue (200 seats)
```bash
VITE_VENUE_ROWS=A,B,C,D,E,F,G,H,I,J
VITE_VENUE_SEATS_PER_ROW=20
VITE_VENUE_TOTAL_SEATS=200
```

### Theater Style (100 seats)
```bash
VITE_VENUE_ROWS=AA,BB,CC,DD,EE,FF,GG,HH,II,JJ
VITE_VENUE_SEATS_PER_ROW=10
VITE_VENUE_TOTAL_SEATS=100
```

## Notes

- All environment variables starting with `VITE_` are available in the browser
- Changes to environment variables require a restart of the development server
- The backend must be configured to match the total number of seats configured in the UI
- Seat numbering follows the pattern: S001, S002, S003, etc.
- Row indexing starts from 0 (A=0, B=1, C=2, etc.)

## Security Considerations

- Never put sensitive information in environment variables that start with `VITE_`
- In production, implement proper user authentication instead of using `VITE_DEFAULT_USER_ID`
- Consider using a proper configuration management system for production deployments 