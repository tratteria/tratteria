import { Injectable } from '@angular/core';
import { BehaviorSubject } from 'rxjs';

@Injectable({ providedIn: 'root' })
export class ModalService {
  private modalMessage = new BehaviorSubject<string>('');

  constructor() {}

  open(message: string) {
    this.modalMessage.next(message);
  }

  close() {
    this.modalMessage.next('');
  }

  getMessage() {
    return this.modalMessage.asObservable();
  }
}
