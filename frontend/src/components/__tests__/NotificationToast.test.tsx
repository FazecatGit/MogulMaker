import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import NotificationToast from '../NotificationToast';
import { useGlobalStore } from '../../store/useGlobalStore';

// Mock the global store
jest.mock('@/store/useGlobalStore');

const mockMarkNotificationRead = jest.fn();

describe('NotificationToast', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    jest.useFakeTimers();
    (useGlobalStore as unknown as jest.Mock).mockImplementation((selector) =>
      selector({
        notifications: [],
        markNotificationRead: mockMarkNotificationRead,
      })
    );
  });

  afterEach(() => {
    jest.runOnlyPendingTimers();
    jest.useRealTimers();
  });

  it('renders nothing when no notifications', () => {
    const { container } = render(<NotificationToast />);
    expect(container.firstChild).toBeNull();
  });

  it('renders success notification with correct styling', () => {
    (useGlobalStore as unknown as jest.Mock).mockImplementation((selector) =>
      selector({
        notifications: [
          {
            id: '1',
            title: 'Success!',
            message: 'Operation completed',
            type: 'success',
            timestamp: Date.now(),
            read: false,
          },
        ],
        markNotificationRead: mockMarkNotificationRead,
      })
    );

    render(<NotificationToast />);
    expect(screen.getByText('Success!')).toBeInTheDocument();
    expect(screen.getByText('Operation completed')).toBeInTheDocument();
  });

  it('renders error notification with correct icon', () => {
    (useGlobalStore as unknown as jest.Mock).mockImplementation((selector) =>
      selector({
        notifications: [
          {
            id: '1',
            title: 'Error',
            message: 'Something went wrong',
            type: 'error',
            timestamp: Date.now(),
            read: false,
          },
        ],
        markNotificationRead: mockMarkNotificationRead,
      })
    );

    const { container } = render(<NotificationToast />);
    expect(screen.getByText('Error')).toBeInTheDocument();
    // Check for red styling
    const notification = container.querySelector('.bg-red-900\\/90');
    expect(notification).toBeInTheDocument();
  });

  it('dismisses notification when close button is clicked', async () => {
    (useGlobalStore as unknown as jest.Mock).mockImplementation((selector) =>
      selector({
        notifications: [
          {
            id: '1',
            title: 'Info',
            message: 'Test message',
            type: 'info',
            timestamp: Date.now(),
            read: false,
          },
        ],
        markNotificationRead: mockMarkNotificationRead,
      })
    );

    const user = userEvent.setup({ delay: null });
    render(<NotificationToast />);
    const closeButton = screen.getByRole('button');
    await user.click(closeButton);

    expect(mockMarkNotificationRead).toHaveBeenCalledWith('1');
  });

  it('auto-dismisses notification after 5 seconds', () => {
    const timestamp = Date.now();
    (useGlobalStore as unknown as jest.Mock).mockImplementation((selector) =>
      selector({
        notifications: [
          {
            id: '1',
            title: 'Info',
            message: 'Auto dismiss test',
            type: 'info',
            timestamp,
            read: false,
          },
        ],
        markNotificationRead: mockMarkNotificationRead,
      })
    );

    render(<NotificationToast />);
    
    // Fast-forward time by 5 seconds
    jest.advanceTimersByTime(5000);

    expect(mockMarkNotificationRead).toHaveBeenCalledWith('1');
  });

  it('only shows up to 3 notifications', () => {
    (useGlobalStore as unknown as jest.Mock).mockImplementation((selector) =>
      selector({
        notifications: [
          { id: '1', title: 'First', message: 'msg1', type: 'info', timestamp: Date.now(), read: false },
          { id: '2', title: 'Second', message: 'msg2', type: 'info', timestamp: Date.now(), read: false },
          { id: '3', title: 'Third', message: 'msg3', type: 'info', timestamp: Date.now(), read: false },
          { id: '4', title: 'Fourth', message: 'msg4', type: 'info', timestamp: Date.now(), read: false },
        ],
        markNotificationRead: mockMarkNotificationRead,
      })
    );

    const { container } = render(<NotificationToast />);
    const notifications = container.querySelectorAll('.animate-slide-in');
    expect(notifications).toHaveLength(3);
  });

  it('filters out read notifications', () => {
    (useGlobalStore as unknown as jest.Mock).mockImplementation((selector) =>
      selector({
        notifications: [
          { id: '1', title: 'Unread', message: 'msg1', type: 'info', timestamp: Date.now(), read: false },
          { id: '2', title: 'Read', message: 'msg2', type: 'info', timestamp: Date.now(), read: true },
        ],
        markNotificationRead: mockMarkNotificationRead,
      })
    );

    render(<NotificationToast />);
    expect(screen.getByText('Unread')).toBeInTheDocument();
    expect(screen.queryByText('Read')).not.toBeInTheDocument();
  });
});
