// Copyright (c) 2025 Lazycat Apps
// Licensed under the MIT License. See LICENSE file in the project root for details.

import { render, screen } from '@testing-library/react';
import App from './App';

const createMockResponse = (data, ok = true) => ({
  ok,
  json: () => Promise.resolve(data),
  blob: () => Promise.resolve(new Blob()),
});

beforeEach(() => {
  global.fetch = jest.fn((input) => {
    const url = typeof input === 'string' ? input : input.url;

    if (url.includes('/api/v1/auth/userinfo')) {
      return Promise.resolve(
        createMockResponse({
          authenticated: false,
          oidc_enabled: false,
        }),
      );
    }

    if (url.includes('/api/v1/quota')) {
      return Promise.resolve(
        createMockResponse({
          quota: {
            userId: 'test',
            totalBytes: 1024,
            usedBytes: 0,
            percentage: 0,
            clientsCount: 0,
            maxClients: 1000,
            updatedAt: new Date().toISOString(),
          },
        }),
      );
    }

    if (url.includes('/api/v1/clients')) {
      return Promise.resolve(
        createMockResponse({
          total: 0,
          page: 1,
          pageSize: 20,
          clients: [],
        }),
      );
    }

    return Promise.resolve(createMockResponse({}));
  });

  if (!window.matchMedia) {
    window.matchMedia = jest.fn();
  }
  window.matchMedia.mockImplementation((query) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: jest.fn(),
    removeListener: jest.fn(),
    addEventListener: jest.fn(),
    removeEventListener: jest.fn(),
    dispatchEvent: jest.fn(),
  }));
});

test('renders gosmee web ui title', async () => {
  render(<App />);
  expect(await screen.findByText(/Gosmee Web UI/)).toBeInTheDocument();
});
