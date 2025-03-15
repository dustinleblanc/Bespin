# Bespin Web Frontend

A modern web interface for the Bespin application, built with Nuxt 3, Vue 3, and TypeScript.

## Features

- 🚀 Built with [Nuxt 3](https://nuxt.com) and Vue 3
- 💪 TypeScript support
- 🎨 TailwindCSS for styling
- 📦 Pinia for state management
- 🔄 Real-time updates with Socket.IO
- ✅ Comprehensive testing suite with Vitest

## Setup

Make sure to install dependencies:

```bash
# pnpm (recommended)
pnpm install
```

## Development

Start the development server on `http://localhost:3000`:

```bash
pnpm dev
```

## Testing

We use Vitest for testing, providing a modern and fast testing experience that's fully compatible with our Vue 3 and TypeScript setup.

### Running Tests

```bash
# Run tests once and exit
pnpm test

# Run tests in watch mode (for development)
pnpm test:watch

# Run tests with UI
pnpm test:ui

# Generate test coverage report
pnpm test:coverage
```

### Testing Philosophy

Our testing approach focuses on:

1. **Component Testing**: Using `@testing-library/vue` for component tests that simulate real user interactions
2. **Store Testing**: Testing Pinia stores for state management logic
3. **Integration Testing**: Testing component interactions and data flow
4. **Best Practices**:
   - Testing user interactions over implementation details
   - Using accessible queries from Testing Library
   - Maintaining test isolation
   - Following the component-driven development approach

### Example Test

```typescript
import { render, screen } from '@testing-library/vue'
import YourComponent from './YourComponent.vue'

describe('YourComponent', () => {
  it('should render correctly', () => {
    render(YourComponent, {
      props: {
        title: 'Hello'
      }
    })
    expect(screen.getByText('Hello')).toBeTruthy()
  })
})
```

## Project Structure

```
web/
├── components/     # Vue components
├── stores/         # Pinia stores
├── test/          # Test setup and utilities
├── public/        # Static assets
├── server/        # Server middleware
└── app.vue        # App root component
```

## Production

Build the application for production:

```bash
pnpm build
```

Preview the production build:

```bash
pnpm preview
```

## Development Tools

- [Nuxt DevTools](https://devtools.nuxtjs.org/)
- [Vue DevTools](https://devtools.vuejs.org/)
- [Vitest UI](https://vitest.dev/guide/ui.html)

## Contributing

1. Write tests for new features
2. Ensure all tests pass: `pnpm test`
3. Generate and review test coverage: `pnpm test:coverage`
4. Follow the existing code style and component patterns

## Environment Variables

Create a `.env` file in the root directory:

```env
API_BASE_URL=http://localhost:8080
SOCKET_URL=http://localhost:8080
```

## Dependencies

- Nuxt 3
- Vue 3
- Pinia for state management
- TailwindCSS for styling
- Socket.IO for real-time communication
- Vitest for testing
- Testing Library for component testing

## Learn More

- [Nuxt 3 Documentation](https://nuxt.com/docs)
- [Vue 3 Documentation](https://vuejs.org/)
- [Vitest Documentation](https://vitest.dev/)
- [Testing Library Documentation](https://testing-library.com/docs/vue-testing-library/intro/)
- [Pinia Documentation](https://pinia.vuejs.org/)
