# ---- Base Stage ----
# Use an official Node.js runtime as a parent image
# Choose an LTS version (e.g., 20) on Alpine for smaller size
FROM node:20-alpine AS base
WORKDIR /usr/src/app

# ---- Dependencies Stage ----
FROM base AS dependencies
# Copy package.json and package-lock.json (or yarn.lock)
COPY package*.json ./
# Install app dependencies using npm ci (clean install, usually faster and safer for CI/CD)
RUN npm ci --only=production
# If you have native dependencies that need build tools:
# RUN apk add --no-cache --virtual .build-deps make gcc g++ python3 && \
#     npm ci --only=production && \
#     apk del .build-deps

# ---- Build Stage (if using TypeScript or a build step) ----
# If you were using TypeScript, you'd have a build stage here:
# FROM base AS build
# COPY --from=dependencies /usr/src/app/node_modules /usr/src/app/node_modules
# COPY . .
# RUN npm run build # Assuming a build script in package.json

# ---- Runtime Stage ----
FROM base AS runtime
# Copy only necessary production dependencies from the 'dependencies' stage
COPY --from=dependencies /usr/src/app/node_modules ./node_modules
# Copy application code (or built code from 'build' stage if applicable)
COPY . .
# If using a build step, copy from the 'build' stage instead:
# COPY --from=build /usr/src/app/dist ./dist # Example if output is in 'dist'

# Expose the port the app runs on
EXPOSE 8082

# Define user to run the application (security best practice)
# Create a non-root user and group called 'appuser'
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

# Define the command to run your app
# Use node directly to execute the main server file
CMD ["node", "server.js"]
# If using a build step and output is in 'dist':
# CMD ["node", "dist/server.js"]