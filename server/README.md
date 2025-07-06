# Server Configuration

The ticket booking server supports dynamic configuration through environment variables. This allows you to customize the application behavior without modifying the code.

## Environment Variables

### Server Configuration
- `PORT` - Server port (default: `8080`)
- `READ_TIMEOUT` - HTTP read timeout (default: `15s`)
- `WRITE_TIMEOUT` - HTTP write timeout (default: `15s`)
- `IDLE_TIMEOUT` - HTTP idle timeout (default: `60s`)

### Database Configuration
- `DB_HOST` - Database host (default: `localhost`)
- `DB_PORT` - Database port (default: `5432`)
- `DB_USER` - Database username (default: `postgres`)
- `DB_PASSWORD` - Database password (default: `password`)
- `DB_NAME` - Database name (default: `ticket_booking`)
- `DB_SSL_MODE` - SSL mode for database connection (default: `disable`)
- `DB_MAX_OPEN_CONNS` - Maximum open database connections (default: `25`)
- `DB_MAX_IDLE_CONNS` - Maximum idle database connections (default: `5`)
- `DB_CONN_MAX_LIFETIME` - Maximum lifetime for database connections (default: `5m`)

### Application Configuration
- `LOG_LEVEL` - Logging level: `debug`, `info`, `warn`, `error` (default: `info`)
- `RATE_LIMIT_RPS` - Rate limiting requests per second (default: `100`)
- `LOCK_TIMEOUT` - General lock timeout for operations (default: `30s`)
- `MAX_RETRIES` - Maximum retries for failed operations (default: `3`)
- `RETRY_DELAY` - Delay between retries (default: `100ms`)

### Seat Locking and Booking Configuration
- `SEAT_LOCK_DURATION` - How long seats remain locked during selection (default: `3m`)
- `BOOKING_EXPIRATION` - How long users have to complete payment after booking (default: `15m`)
- `CLEANUP_INTERVAL` - How often to run cleanup routine for expired seat locks (default: `1m`)

## Duration Format

Duration values support Go's duration format:
- `s` - seconds (e.g., `30s`)
- `m` - minutes (e.g., `5m`)
- `h` - hours (e.g., `2h`)
- `ms` - milliseconds (e.g., `100ms`)

Examples:
- `SEAT_LOCK_DURATION=5m` - Lock seats for 5 minutes
- `BOOKING_EXPIRATION=30m` - Give users 30 minutes to complete payment
- `CLEANUP_INTERVAL=30s` - Run cleanup every 30 seconds

## Example Configuration

Create a `.env` file in the server directory:

```bash
# Server Configuration
PORT=8080
READ_TIMEOUT=15s
WRITE_TIMEOUT=15s
IDLE_TIMEOUT=60s

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_secure_password
DB_NAME=ticket_booking
DB_SSL_MODE=disable

# Application Configuration
LOG_LEVEL=info
RATE_LIMIT_RPS=100

# Seat Locking Configuration
SEAT_LOCK_DURATION=3m      # Lock seats for 3 minutes during selection
BOOKING_EXPIRATION=15m     # Users have 15 minutes to complete payment
CLEANUP_INTERVAL=1m        # Clean up expired locks every minute
```

## Configuration Examples

### High-Traffic Environment
```bash
# Increase connection limits and adjust timeouts
DB_MAX_OPEN_CONNS=50
DB_MAX_IDLE_CONNS=10
RATE_LIMIT_RPS=500

# Faster cleanup for high turnover
SEAT_LOCK_DURATION=2m
CLEANUP_INTERVAL=30s
```

### Development Environment
```bash
# More verbose logging
LOG_LEVEL=debug

# Longer timeouts for debugging
SEAT_LOCK_DURATION=10m
BOOKING_EXPIRATION=60m

# Less frequent cleanup
CLEANUP_INTERVAL=2m
```

### Production Environment
```bash
# Secure database connection
DB_SSL_MODE=require
DB_PASSWORD=your_very_secure_password

# Production logging
LOG_LEVEL=warn

# Balanced settings
SEAT_LOCK_DURATION=3m
BOOKING_EXPIRATION=15m
CLEANUP_INTERVAL=1m
```

## Seat Locking Behavior

The seat locking system works as follows:

1. **User selects a seat**: Seat status changes to `locked` for the duration specified by `SEAT_LOCK_DURATION`
2. **Lock expiration**: If user doesn't complete booking within the lock duration, seat becomes available again
3. **Cleanup routine**: Runs every `CLEANUP_INTERVAL` to release expired locks
4. **Booking creation**: When user proceeds to payment, locked seats become `reserved`
5. **Payment timeout**: Users have `BOOKING_EXPIRATION` time to complete payment
6. **Final status**: After successful payment, seats become `sold`

## Performance Considerations

### Seat Lock Duration
- **Too short** (< 1m): Users may lose seats while filling out payment forms
- **Too long** (> 10m): Seats may be unavailable for too long, reducing sales
- **Recommended**: 3-5 minutes for most use cases

### Cleanup Interval
- **Too frequent** (< 30s): Unnecessary database load
- **Too infrequent** (> 5m): Seats may appear unavailable longer than necessary
- **Recommended**: 1-2 minutes for most use cases

### Booking Expiration
- **Too short** (< 5m): Users may not have enough time for payment processing
- **Too long** (> 30m): Reserved seats unavailable for too long
- **Recommended**: 10-20 minutes depending on payment provider

## Security Considerations

- Always use strong passwords for database connections
- Enable SSL for database connections in production (`DB_SSL_MODE=require`)
- Use appropriate rate limiting based on your expected traffic
- Monitor logs regularly (set appropriate `LOG_LEVEL`)
- Consider using environment-specific configuration files

## Monitoring

The server logs important events including:
- Seat lock/unlock operations
- Booking creation and confirmation
- Cleanup operations with statistics
- Configuration values at startup

Example log output:
```json
{
  "level": "info",
  "msg": "Cleaned up expired seat locks",
  "seats_unlocked": 5,
  "lock_duration": "3m0s",
  "time": "2024-01-01T12:00:00Z"
}
```

## Troubleshooting

### Common Issues

1. **Seats stuck in locked state**
   - Check `CLEANUP_INTERVAL` is running (should see logs every interval)
   - Verify `SEAT_LOCK_DURATION` is appropriate
   - Check for database connectivity issues

2. **Bookings expiring too quickly**
   - Increase `BOOKING_EXPIRATION` duration
   - Check payment provider integration timeouts

3. **High database load**
   - Increase `CLEANUP_INTERVAL` to reduce frequency
   - Optimize database connection settings
   - Consider adding database indexes

4. **Rate limiting issues**
   - Adjust `RATE_LIMIT_RPS` based on traffic patterns
   - Monitor rate limiting logs

For more detailed troubleshooting, enable debug logging:
```bash
LOG_LEVEL=debug
``` 