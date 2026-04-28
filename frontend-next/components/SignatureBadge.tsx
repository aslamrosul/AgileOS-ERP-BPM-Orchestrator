'use client';

import { useState } from 'react';
import { QRCodeSVG } from 'qrcode.react';
import { CheckCircle, Shield, Eye, Copy } from 'lucide-react';

interface SignatureBadgeProps {
  taskId: string;
  signature: string;
  signedBy: string;
  signedAt: string;
  qrCodeData: string;
  stepName?: string;
  onVerify?: (taskId: string, signature: string) => void;
}

export function SignatureBadge({
  taskId,
  signature,
  signedBy,
  signedAt,
  qrCodeData,
  stepName,
  onVerify
}: SignatureBadgeProps) {
  const [showFullSignature, setShowFullSignature] = useState(false);
  const [copied, setCopied] = useState(false);

  const handleCopySignature = async () => {
    try {
      await navigator.clipboard.writeText(signature);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error('Failed to copy signature:', err);
    }
  };

  const handleVerify = () => {
    if (onVerify) {
      onVerify(taskId, signature);
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  const truncateSignature = (sig: string, length: number = 16) => {
    return sig.length > length ? `${sig.substring(0, length)}...` : sig;
  };

  return (
    <div className="signature-badge bg-white border border-green-200 rounded-lg p-6 shadow-sm">
      {/* Header */}
      <div className="flex items-center gap-3 mb-4">
        <div className="flex items-center justify-center w-10 h-10 bg-green-100 rounded-full">
          <CheckCircle className="w-6 h-6 text-green-600" />
        </div>
        <div>
          <h3 className="text-lg font-semibold text-gray-900">Digitally Signed</h3>
          {stepName && (
            <p className="text-sm text-gray-600">{stepName}</p>
          )}
        </div>
        <div className="ml-auto">
          <Shield className="w-6 h-6 text-green-600" />
        </div>
      </div>

      {/* Signature Details */}
      <div className="space-y-3 mb-4">
        <div className="flex justify-between items-center">
          <span className="text-sm font-medium text-gray-700">Signed by:</span>
          <span className="text-sm text-gray-900">{signedBy}</span>
        </div>
        
        <div className="flex justify-between items-center">
          <span className="text-sm font-medium text-gray-700">Date:</span>
          <span className="text-sm text-gray-900">{formatDate(signedAt)}</span>
        </div>
        
        <div className="flex justify-between items-start">
          <span className="text-sm font-medium text-gray-700">Signature:</span>
          <div className="flex items-center gap-2">
            <code className="text-xs font-mono bg-gray-100 px-2 py-1 rounded">
              {showFullSignature ? signature : truncateSignature(signature)}
            </code>
            <button
              onClick={() => setShowFullSignature(!showFullSignature)}
              className="text-blue-600 hover:text-blue-800"
              title={showFullSignature ? "Hide full signature" : "Show full signature"}
            >
              <Eye className="w-4 h-4" />
            </button>
            <button
              onClick={handleCopySignature}
              className="text-gray-600 hover:text-gray-800"
              title="Copy signature"
            >
              <Copy className="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>

      {/* QR Code Section */}
      <div className="flex items-center gap-4 mb-4">
        <div className="flex-shrink-0">
          <QRCodeSVG 
            value={qrCodeData} 
            size={80}
            level="M"
            includeMargin={true}
            className="border border-gray-200 rounded"
          />
        </div>
        <div className="flex-1">
          <p className="text-sm font-medium text-gray-700 mb-1">Verification QR Code</p>
          <p className="text-xs text-gray-600">
            Scan to verify signature authenticity and task integrity
          </p>
        </div>
      </div>

      {/* Action Buttons */}
      <div className="flex gap-2">
        <button
          onClick={handleVerify}
          className="flex-1 bg-blue-600 text-white px-4 py-2 rounded-md text-sm font-medium hover:bg-blue-700 transition-colors"
        >
          Verify Signature
        </button>
        <button
          onClick={() => window.open(qrCodeData, '_blank')}
          className="px-4 py-2 border border-gray-300 text-gray-700 rounded-md text-sm font-medium hover:bg-gray-50 transition-colors"
        >
          View Details
        </button>
      </div>

      {/* Copy Feedback */}
      {copied && (
        <div className="mt-2 text-xs text-green-600 text-center">
          Signature copied to clipboard!
        </div>
      )}

      {/* Security Notice */}
      <div className="mt-4 p-3 bg-blue-50 border border-blue-200 rounded-md">
        <p className="text-xs text-blue-800">
          <Shield className="w-3 h-3 inline mr-1" />
          This digital signature provides cryptographic proof that this approval is authentic and has not been tampered with.
        </p>
      </div>
    </div>
  );
}

// Verification Modal Component
interface VerificationModalProps {
  isOpen: boolean;
  onClose: () => void;
  verificationResult: any;
}

export function VerificationModal({ isOpen, onClose, verificationResult }: VerificationModalProps) {
  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4">
        <div className="flex items-center gap-3 mb-4">
          {verificationResult?.valid ? (
            <CheckCircle className="w-8 h-8 text-green-600" />
          ) : (
            <Shield className="w-8 h-8 text-red-600" />
          )}
          <h2 className="text-xl font-semibold">
            Signature Verification
          </h2>
        </div>

        <div className="space-y-3 mb-6">
          <div className="flex justify-between">
            <span className="font-medium">Status:</span>
            <span className={`font-semibold ${
              verificationResult?.valid ? 'text-green-600' : 'text-red-600'
            }`}>
              {verificationResult?.valid ? 'Valid' : 'Invalid'}
            </span>
          </div>
          
          <div className="flex justify-between">
            <span className="font-medium">Verified at:</span>
            <span className="text-sm">
              {verificationResult?.verified_at ? 
                new Date(verificationResult.verified_at).toLocaleString() : 
                'N/A'
              }
            </span>
          </div>
          
          <div className="mt-4">
            <p className="text-sm text-gray-700">
              {verificationResult?.message || 'No verification message available'}
            </p>
          </div>
        </div>

        <button
          onClick={onClose}
          className="w-full bg-gray-600 text-white px-4 py-2 rounded-md hover:bg-gray-700 transition-colors"
        >
          Close
        </button>
      </div>
    </div>
  );
}