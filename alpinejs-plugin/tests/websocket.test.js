// Mock src/index.js to prevent its side effects during websocket tests
jest.mock('../src/index.js', () => ({
    __esModule: true,
    default: jest.fn(),
}))

// Define the mock function instances at the top level of your test suite.
// These will be the actual jest.fn() instances you control and assert against.
const mockGetSessionIDFromCookieFn = jest.fn()
const mockDefaultCreateWebSocketFn = jest.fn()

describe('setupWebSocketConnection', () => {
    let mockFetch
    let mockProcessEventsFn
    let originalWindowLocation
    let setupWebSocketConnection // Will hold the dynamically imported function

    beforeEach(async () => {
        jest.resetModules() // Reset the module cache before each test

        // Use jest.doMock to set up mocks. These will be applied when
        // '../src/websocket' (and its dependencies) are imported next.
        jest.doMock('../src/utils', () => ({
            __esModule: true,
            getSessionIDFromCookie: mockGetSessionIDFromCookieFn,
        }))

        jest.doMock('../src/websocket', () => {
            // Important: requireActual must be inside the factory for jest.doMock
            // when you want to partially mock a module.
            const originalModule = jest.requireActual('../src/websocket')
            return {
                ...originalModule, // Keep named exports like setupWebSocketConnection
                __esModule: true,
                default: mockDefaultCreateWebSocketFn, // Mock the default export (createWebSocket)
            }
        })

        // Dynamically import the module under test AFTER mocks are in place.
        // This ensures it gets the mocked dependencies.
        const wsModule = await import('../src/websocket')
        setupWebSocketConnection = wsModule.setupWebSocketConnection

        // Reset the state of the mock functions for the current test
        mockGetSessionIDFromCookieFn.mockReset()
        mockDefaultCreateWebSocketFn.mockReset()

        // Standard test setup
        mockProcessEventsFn = jest.fn()
        mockFetch = jest.fn()
        global.fetch = mockFetch

        originalWindowLocation = window.location
        delete window.location
        window.location = {
            host: 'localhost:8080',
            pathname: '/testpath',
            protocol: 'http:',
            href: 'http://localhost:8080/testpath',
        }

        console.error = jest.fn()
        console.log = jest.fn()
    })

    afterEach(() => {
        window.location = originalWindowLocation
    })

    test('should establish WebSocket connection when enabled and session ID exists', async () => {
        mockGetSessionIDFromCookieFn.mockReturnValue('test-session-id')
        mockFetch.mockResolvedValue({
            headers: { get: jest.fn().mockReturnValue('true') },
        })
        const mockWsInstance = {
            connect: jest.fn(),
            emit: jest.fn(),
            close: jest.fn(),
            getState: jest.fn(),
        }
        mockDefaultCreateWebSocketFn.mockReturnValue(mockWsInstance)

        const ws = await setupWebSocketConnection(
            mockProcessEventsFn,
            mockFetch,
            window.location,
            mockDefaultCreateWebSocketFn // Explicitly pass the mock factory
        )

        expect(mockGetSessionIDFromCookieFn).toHaveBeenCalled()
        expect(mockFetch).toHaveBeenCalledWith(
            'http://localhost:8080/testpath',
            { method: 'HEAD' }
        )
        expect(mockDefaultCreateWebSocketFn).toHaveBeenCalledWith(
            'ws://localhost:8080/testpath',
            [],
            expect.any(Function)
        )
        expect(ws).toBe(mockWsInstance)
        expect(console.log).toHaveBeenCalledWith(
            'WebSocket enabled, attempting connection...'
        )
    })

    test('should use wss protocol when window.location.protocol is https:', async () => {
        window.location.protocol = 'https:'
        window.location.href = 'https://localhost:8080/testpath'
        mockGetSessionIDFromCookieFn.mockReturnValue('test-session-id')
        mockFetch.mockResolvedValue({
            headers: { get: jest.fn().mockReturnValue('true') },
        })
        const mockWsInstance = {
            connect: jest.fn(),
            emit: jest.fn(),
            close: jest.fn(),
            getState: jest.fn(),
        }
        mockDefaultCreateWebSocketFn.mockReturnValue(mockWsInstance)

        await setupWebSocketConnection(
            mockProcessEventsFn,
            mockFetch,
            window.location,
            mockDefaultCreateWebSocketFn // Explicitly pass the mock factory
        )

        expect(mockDefaultCreateWebSocketFn).toHaveBeenCalledWith(
            'wss://localhost:8080/testpath',
            [],
            expect.any(Function)
        )
    })

    test('should use explicitly passed wsFactory if provided', async () => {
        mockGetSessionIDFromCookieFn.mockReturnValue('test-session-id')
        mockFetch.mockResolvedValue({
            headers: { get: jest.fn().mockReturnValue('true') },
        })
        const explicitMockWsInstance = { customConnect: jest.fn() }
        const explicitUserProvidedWsFactory = jest
            .fn()
            .mockReturnValue(explicitMockWsInstance)

        const ws = await setupWebSocketConnection(
            mockProcessEventsFn,
            mockFetch,
            window.location,
            explicitUserProvidedWsFactory // This test already does it correctly
        )

        expect(explicitUserProvidedWsFactory).toHaveBeenCalledWith(
            'ws://localhost:8080/testpath',
            [],
            expect.any(Function)
        )
        expect(mockDefaultCreateWebSocketFn).not.toHaveBeenCalled()
        expect(ws).toBe(explicitMockWsInstance)
    })

    test('should not connect if WebSocket is not enabled by server', async () => {
        mockGetSessionIDFromCookieFn.mockReturnValue('test-session-id')
        mockFetch.mockResolvedValue({
            headers: { get: jest.fn().mockReturnValue('false') },
        })

        const ws = await setupWebSocketConnection(
            mockProcessEventsFn,
            mockFetch,
            window.location
            // No factory passed, as it's not expected to be called
        )

        expect(mockGetSessionIDFromCookieFn).toHaveBeenCalled()
        expect(mockDefaultCreateWebSocketFn).not.toHaveBeenCalled() // This assertion remains key
        expect(ws).toBeNull()
        expect(console.log).toHaveBeenCalledWith(
            'WebSocket not enabled by server.'
        )
    })

    test('should not connect if session ID is missing', async () => {
        mockGetSessionIDFromCookieFn.mockReturnValue(null)

        const ws = await setupWebSocketConnection(
            mockProcessEventsFn,
            mockFetch,
            window.location
            // No factory passed, as it's not expected to be called
        )

        expect(mockGetSessionIDFromCookieFn).toHaveBeenCalled()
        expect(mockFetch).not.toHaveBeenCalled()
        expect(mockDefaultCreateWebSocketFn).not.toHaveBeenCalled() // This assertion remains key
        expect(ws).toBeNull()
        expect(console.error).toHaveBeenCalledWith(
            'No session ID found in cookie. WebSocket disabled.'
        )
    })

    test('should handle errors during HEAD request', async () => {
        mockGetSessionIDFromCookieFn.mockReturnValue('test-session-id')
        const headError = new Error('HEAD request failed')
        mockFetch.mockRejectedValue(headError)

        const ws = await setupWebSocketConnection(
            mockProcessEventsFn,
            mockFetch,
            window.location
            // No factory passed, as it's not expected to be called
        )

        expect(mockGetSessionIDFromCookieFn).toHaveBeenCalled()
        expect(mockFetch).toHaveBeenCalled()
        expect(mockDefaultCreateWebSocketFn).not.toHaveBeenCalled() // This assertion remains key
        expect(ws).toBeNull()
        expect(console.error).toHaveBeenCalledWith(
            'Error checking WebSocket status:',
            headError
        )
    })
})
