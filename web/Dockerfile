FROM node:20-slim

# Enable pnpm
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable

WORKDIR /app

# Copy the entire application
COPY . .

# Install dependencies
RUN pnpm install

# Start the development server
CMD ["pnpm", "dev"]
