import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/task_provider.dart';

class ErrorScreen extends ConsumerWidget {
  final String error;
  
  const ErrorScreen({super.key, required this.error});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return Scaffold(
      body: Center(
        child: Padding(
          padding: const EdgeInsets.all(20),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(
                Icons.error_outline,
                size: 80,
                color: Theme.of(context).colorScheme.error,
              ),
              const SizedBox(height: 20),
              Text(
                'Something went wrong',
                style: Theme.of(context).textTheme.titleLarge?.copyWith(
                  color: Theme.of(context).colorScheme.error,
                ),
              ),
              const SizedBox(height: 10),
              Text(
                error,
                textAlign: TextAlign.center,
                style: Theme.of(context).textTheme.bodyMedium,
              ),
              const SizedBox(height: 30),
              ElevatedButton(
                onPressed: () => ref.invalidate(pendingTasksProvider),
                child: const Text('Try Again'),
              ),
              const SizedBox(height: 10),
              TextButton(
                onPressed: () {
                  // Show connection help
                  showDialog(
                    context: context,
                    builder: (context) => AlertDialog(
                      title: const Text('Connection Help'),
                      content: const Column(
                        mainAxisSize: MainAxisSize.min,
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text('Make sure:'),
                          SizedBox(height: 8),
                          Text('1. Backend is running on port 8080'),
                          Text('2. Database is seeded with workflows'),
                          Text('3. Correct IP address is configured'),
                          SizedBox(height: 8),
                          Text('For Android emulator: 10.0.2.2'),
                          Text('For iOS simulator: localhost'),
                          Text('For physical device: Your computer IP'),
                        ],
                      ),
                      actions: [
                        TextButton(
                          onPressed: () => Navigator.pop(context),
                          child: const Text('OK'),
                        ),
                      ],
                    ),
                  );
                },
                child: const Text('Connection Help'),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
