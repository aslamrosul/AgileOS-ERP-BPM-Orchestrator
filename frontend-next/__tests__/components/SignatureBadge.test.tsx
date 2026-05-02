import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import SignatureBadge from '@/components/SignatureBadge';

// Mock qrcode.react
vi.mock('qrcode.react', () => ({
  QRCodeSVG: ({ value }: { value: string }) => (
    <svg data-testid="qr-code" data-value={value}>
      QR Code Mock
    </svg>
  ),
}));

describe('SignatureBadge Component', () => {
  const mockSignature = {
    task_id: 'task_123',
    signature: 'abc123def456',
    timestamp: '2024-01-15T10:30:00Z',
    user_id: 'user_456',
    username: 'John Doe',
  };

  it('renders signature badge with QR code when data is provided', () => {
    render(<SignatureBadge signature={mockSignature} />);

    // Check if QR code is rendered
    const qrCode = screen.getByTestId('qr-code');
    expect(qrCode).toBeInTheDocument();
    expect(qrCode).toHaveAttribute('data-value', mockSignature.signature);
  });

  it('displays signature hash', () => {
    render(<SignatureBadge signature={mockSignature} />);

    // Check if signature hash is displayed (truncated)
    expect(screen.getByText(/abc123/)).toBeInTheDocument();
  });

  it('displays username', () => {
    render(<SignatureBadge signature={mockSignature} />);

    expect(screen.getByText('John Doe')).toBeInTheDocument();
  });

  it('displays formatted timestamp', () => {
    render(<SignatureBadge signature={mockSignature} />);

    // Check if timestamp is displayed (format may vary)
    const timestampElement = screen.getByText(/2024/);
    expect(timestampElement).toBeInTheDocument();
  });

  it('renders nothing when signature is null', () => {
    const { container } = render(<SignatureBadge signature={null} />);

    expect(container.firstChild).toBeNull();
  });

  it('renders nothing when signature is undefined', () => {
    const { container } = render(<SignatureBadge signature={undefined} />);

    expect(container.firstChild).toBeNull();
  });

  it('displays task ID', () => {
    render(<SignatureBadge signature={mockSignature} />);

    expect(screen.getByText(/task_123/)).toBeInTheDocument();
  });

  it('applies correct CSS classes for styling', () => {
    const { container } = render(<SignatureBadge signature={mockSignature} />);

    // Check if container has expected classes
    const badge = container.firstChild as HTMLElement;
    expect(badge).toHaveClass('signature-badge');
  });

  it('truncates long signature hashes', () => {
    const longSignature = {
      ...mockSignature,
      signature: 'a'.repeat(100),
    };

    render(<SignatureBadge signature={longSignature} />);

    // Should display truncated version
    const signatureText = screen.getByText(/aaa/);
    expect(signatureText.textContent).not.toEqual(longSignature.signature);
    expect(signatureText.textContent?.length).toBeLessThan(longSignature.signature.length);
  });

  it('handles missing optional fields gracefully', () => {
    const minimalSignature = {
      task_id: 'task_123',
      signature: 'abc123',
      timestamp: '2024-01-15T10:30:00Z',
    };

    const { container } = render(<SignatureBadge signature={minimalSignature as any} />);

    // Should still render without errors
    expect(container.firstChild).toBeInTheDocument();
    expect(screen.getByTestId('qr-code')).toBeInTheDocument();
  });
});

describe('SignatureBadge Accessibility', () => {
  const mockSignature = {
    task_id: 'task_123',
    signature: 'abc123def456',
    timestamp: '2024-01-15T10:30:00Z',
    user_id: 'user_456',
    username: 'John Doe',
  };

  it('has accessible QR code with alt text', () => {
    render(<SignatureBadge signature={mockSignature} />);

    const qrCode = screen.getByTestId('qr-code');
    expect(qrCode).toHaveAttribute('role', 'img');
  });

  it('provides semantic HTML structure', () => {
    const { container } = render(<SignatureBadge signature={mockSignature} />);

    // Check for semantic elements
    const badge = container.querySelector('.signature-badge');
    expect(badge).toBeInTheDocument();
  });
});

describe('SignatureBadge Integration', () => {
  it('generates correct QR code data', () => {
    const signature = {
      task_id: 'task_123',
      signature: 'test_signature_hash',
      timestamp: '2024-01-15T10:30:00Z',
      user_id: 'user_456',
      username: 'Test User',
    };

    render(<SignatureBadge signature={signature} />);

    const qrCode = screen.getByTestId('qr-code');
    expect(qrCode).toHaveAttribute('data-value', 'test_signature_hash');
  });

  it('updates when signature prop changes', () => {
    const signature1 = {
      task_id: 'task_1',
      signature: 'signature_1',
      timestamp: '2024-01-15T10:30:00Z',
      user_id: 'user_1',
      username: 'User One',
    };

    const signature2 = {
      task_id: 'task_2',
      signature: 'signature_2',
      timestamp: '2024-01-16T10:30:00Z',
      user_id: 'user_2',
      username: 'User Two',
    };

    const { rerender } = render(<SignatureBadge signature={signature1} />);
    expect(screen.getByText('User One')).toBeInTheDocument();

    rerender(<SignatureBadge signature={signature2} />);
    expect(screen.getByText('User Two')).toBeInTheDocument();
    expect(screen.queryByText('User One')).not.toBeInTheDocument();
  });
});