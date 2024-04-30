import { Component, OnInit } from '@angular/core';
import { ModalService } from '../../services/modal.service';

@Component({
  selector: 'app-forbidden-modal',
  templateUrl: './forbidden-modal.component.html',
  styleUrls: ['./forbidden-modal.component.css']
})
export class ForbiddenModalComponent implements OnInit {
  message: string = '';

  constructor(private modalService: ModalService) {}

  ngOnInit() {
    this.modalService.getMessage().subscribe(msg => {
      this.message = msg;
    });
  }

  closeModal() {
    this.modalService.close();
  }
}
